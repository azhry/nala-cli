# nala CLI logging

## Log location

- Default: user config/nala/logs/nala-cli-<pid>.log.
- NALA_LOG_FILE overrides the default path.
- Default path is under the OS user config directory; NALA_CONFIG_DIR changes the config root.
- CLI log entries contain only a recognized command name, outcome, and Go error type. Command arguments, token values, authorization headers, and request bodies are not recorded.

## Rotation

Logs rotate at 10 MiB, keep up to 5 compressed backups, and remove backups older than 30 days. Use a separate file for each concurrently running process; the default CLI filename includes the process ID.

## Diagnosing failures

Inspect the current log and its rotated backups before drawing conclusions from a failure. Check the OS user config directory, NALA_CONFIG_DIR, and NALA_LOG_FILE when the expected file is missing. Treat logs as diagnostic data and do not copy credentials, tokens, or private request data into new log messages.
