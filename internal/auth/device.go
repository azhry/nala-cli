package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const (
	deviceCodeGrantType = "urn:ietf:params:oauth:grant-type:device_code"
	deviceClientID      = "nala-cli"
	defaultDevicePoll   = 5 * time.Second
)

type deviceAuthorizationResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type deviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type deviceOAuthError struct {
	Error string `json:"error"`
}

func (c *Client) LoginDeviceCode(ctx context.Context, output io.Writer) (User, error) {
	if c == nil || c.Store == nil {
		return User{}, errors.New("client is not configured")
	}
	if output == nil {
		return User{}, errors.New("login output is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if c.LoginTimeout <= 0 {
		c.LoginTimeout = defaultLoginTimeout
	}
	loginContext, cancel := context.WithTimeout(ctx, c.LoginTimeout)
	defer cancel()

	authorization, err := c.requestDeviceAuthorization(loginContext)
	if err != nil {
		return User{}, err
	}
	deviceContext, deviceCancel := context.WithTimeout(loginContext, time.Duration(authorization.ExpiresIn)*time.Second)
	defer deviceCancel()
	loginContext = deviceContext
	if _, err := fmt.Fprintf(output, "Visit %s and enter code %s\n", authorization.VerificationURI, authorization.UserCode); err != nil {
		return User{}, fmt.Errorf("write verification instructions: %w", err)
	}

	interval := time.Duration(authorization.Interval) * time.Second
	for {
		if err := c.waitFor(loginContext, interval); err != nil {
			return User{}, errors.New("login timed out or was canceled")
		}

		accessToken, oauthError, err := c.requestDeviceToken(loginContext, authorization.DeviceCode)
		if err != nil {
			if errors.Is(loginContext.Err(), context.DeadlineExceeded) || errors.Is(loginContext.Err(), context.Canceled) {
				return User{}, errors.New("login timed out or was canceled")
			}
			return User{}, err
		}
		switch oauthError {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
			continue
		case "access_denied":
			return User{}, errors.New("device login was denied")
		case "expired_token":
			return User{}, errors.New("device authorization expired")
		case "invalid_grant":
			return User{}, errors.New("device authorization is no longer valid")
		case "":
		default:
			return User{}, errors.New("device authorization failed")
		}

		user, err := c.userInfoWithToken(loginContext, accessToken)
		if err != nil {
			return User{}, fmt.Errorf("validate device session: %w", err)
		}
		if err := c.Store.Save(accessToken); err != nil {
			return User{}, fmt.Errorf("save session: %w", err)
		}
		return user, nil
	}
}

func (c *Client) requestDeviceAuthorization(ctx context.Context) (deviceAuthorizationResponse, error) {
	endpoint, err := c.endpoint("/api/auth/cli/device/code")
	if err != nil {
		return deviceAuthorizationResponse{}, err
	}
	form := url.Values{"client_id": {deviceClientID}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return deviceAuthorizationResponse{}, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.noRedirectClient().Do(request)
	if err != nil {
		return deviceAuthorizationResponse{}, fmt.Errorf("request device authorization: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return deviceAuthorizationResponse{}, fmt.Errorf("device authorization request failed with status %s", response.Status)
	}
	var authorization deviceAuthorizationResponse
	if err := decodeJSON(response.Body, &authorization); err != nil {
		return deviceAuthorizationResponse{}, errors.New("device authorization response is invalid")
	}
	if !validDeviceCode(authorization.DeviceCode) || !validDeviceUserCode(authorization.UserCode) {
		return deviceAuthorizationResponse{}, errors.New("device authorization response is invalid")
	}
	verificationURI, err := validDeviceVerificationURI(authorization.VerificationURI)
	if err != nil {
		return deviceAuthorizationResponse{}, errors.New("device authorization response is invalid")
	}
	if authorization.ExpiresIn <= 0 || authorization.ExpiresIn > 24*60*60 {
		return deviceAuthorizationResponse{}, errors.New("device authorization response is invalid")
	}
	if authorization.Interval <= 0 {
		authorization.Interval = int(defaultDevicePoll.Seconds())
	}
	if authorization.Interval > 60*60 {
		return deviceAuthorizationResponse{}, errors.New("device authorization response is invalid")
	}
	authorization.VerificationURI = verificationURI
	return authorization, nil
}

func (c *Client) requestDeviceToken(ctx context.Context, deviceCode string) (string, string, error) {
	endpoint, err := c.endpoint("/api/auth/cli/device/token")
	if err != nil {
		return "", "", err
	}
	form := url.Values{
		"grant_type":  {deviceCodeGrantType},
		"client_id":   {deviceClientID},
		"device_code": {deviceCode},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.noRedirectClient().Do(request)
	if err != nil {
		return "", "", fmt.Errorf("poll device authorization: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		var result deviceOAuthError
		if err := decodeJSON(response.Body, &result); err != nil || result.Error == "" {
			return "", "", fmt.Errorf("device authorization polling failed with status %s", response.Status)
		}
		return "", result.Error, nil
	}
	var result deviceTokenResponse
	if err := decodeJSON(response.Body, &result); err != nil || !strings.EqualFold(result.TokenType, "Bearer") || !validSessionToken(result.AccessToken) {
		return "", "", errors.New("device token response is invalid")
	}
	return result.AccessToken, "", nil
}

func (c *Client) waitFor(ctx context.Context, duration time.Duration) error {
	if c.Wait != nil {
		return c.Wait(ctx, duration)
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func validDeviceCode(code string) bool {
	if len(code) < 32 || len(code) > 512 {
		return false
	}
	for _, character := range code {
		if !((character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' || character == '-') {
			return false
		}
	}
	return true
}

func validDeviceUserCode(code string) bool {
	if len(code) < 4 || len(code) > 32 {
		return false
	}
	for _, character := range code {
		if !((character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') || (character >= '2' && character <= '9') || character == '-') {
			return false
		}
	}
	return true
}

func validDeviceVerificationURI(raw string) (string, error) {
	if raw == "" || len(raw) > 2048 || strings.TrimSpace(raw) != raw {
		return "", errors.New("invalid verification URI")
	}
	for _, character := range raw {
		if unicode.IsControl(character) {
			return "", errors.New("invalid verification URI")
		}
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" {
		return "", errors.New("invalid verification URI")
	}
	return parsed.String(), nil
}

func validSessionToken(token string) bool {
	return len(token) > 0 && len(token) <= 8192 && strings.Count(token, ".") == 2 && !strings.ContainsAny(token, "\r\n \t")
}
