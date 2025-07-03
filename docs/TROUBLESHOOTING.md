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

**No messages received**: Check ACARSHub connection and port availability -- if you're running ACARSHub in Docker, make sure your docker-compose has the ports open. Ensure you've set ACARSHub to send outputs to ACARS Processor.

**AI filtering not working**: Verify API keys and model availability -- including system status. If you're still having a problem open an issue. 

**Discord webhook failing**: Check webhook URL -- ensure you aren't including any extra characters (you'd be surprised). Ensure you copied the correct webhook. 

**Database errors**: Ensure database file is writable (SQLite) or connection details are correct (MariaDB). Ensure permissions are set correctly. 

**If you run into a problem you can't fix, please open an issue!) 

## Getting help

1. Check logs for error messages
2. [Open an issue](https://github.com/tyzbit/acars-processor/issues) with:
   - Configuration (remove sensitive data like webhook URLs or API keys)
   - Log snippets showing the problem
   - Steps to reproduce
   - What steps you've already tried are a good bonus
   - Your OS + specs
