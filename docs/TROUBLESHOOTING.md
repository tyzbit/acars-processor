# Troubleshooting

## Logs

Enable debug logging:
```yaml
ACARSProcessorSettings:
  LogLevel: debug
```

Check logs for common patterns:
```bash
# Connection issues
grep "Connected to ACARSHub" logs.txt
grep "Failed to connect" logs.txt

# Message processing
grep "Processing message" logs.txt
grep "Message passed all filters" logs.txt

# Errors
grep "ERROR" logs.txt
```

## Configuration

Generate schema for validation:
```bash
./acars-processor -s  # Creates schema.json
```

Test configuration loading:
```bash
./acars-processor -c config.yaml --dry-run
```

## Common issues

**No messages received**: Check ACARSHub connection and port availability.

**AI filtering not working**: Verify API keys and model availability.

**Discord webhook failing**: Check webhook URL and Discord server permissions.

**Database errors**: Ensure database file is writable (SQLite) or connection details are correct (MariaDB).

**High CPU usage**: Reduce filter complexity or enable message rate limiting.

## Getting help

1. Enable debug logging
2. Check logs for error messages
3. [Open an issue](https://github.com/tyzbit/acars-processor/issues) with:
   - Configuration (remove sensitive data)
   - Log snippets showing the problem
   - Steps to reproduce
