# Troubleshooting guide

## Overview

This comprehensive troubleshooting guide addresses common issues, diagnostic procedures, and resolution strategies for ACARS processor deployments. The guide is organized by problem category and includes specific solutions for different deployment environments.

## General diagnostic procedures

### Log analysis

**Enable debug logging for troubleshooting**:
```yaml
ACARSProcessorSettings:
  LogLevel: debug
  ColorOutput: true
```

**Key log patterns to monitor**:
```bash
# Application startup sequence
grep "Connected to ACARSHub" logs.txt
grep "annotators enabled" logs.txt  
grep "receivers configured" logs.txt

# Message processing flow
grep "Processing message" logs.txt
grep "Message passed all filters" logs.txt
grep "Failed to send to receiver" logs.txt

# Error patterns
grep "ERROR" logs.txt
grep "WARN" logs.txt
grep "Failed" logs.txt
```

### Configuration validation

**Schema validation**:
```bash
# Generate current schema
./acars-processor -s

# Validate configuration (requires jsonschema tool)
npm install -g ajv-cli
ajv validate -s schema.json -d config.yaml
```

**Configuration syntax check**:
```bash
# Test YAML parsing
python3 -c "import yaml; yaml.safe_load(open('config.yaml'))"

# Test environment variable substitution
./acars-processor -c config.yaml --dry-run  # Future enhancement
```

### Health checks

**Service health verification**:
```bash
# Check process status
pgrep -f acars-processor
ps aux | grep acars-processor

# Docker container health
docker ps | grep acars-processor
docker logs acars-processor --tail 50

# System resource usage
docker stats acars-processor
top -p $(pgrep acars-processor)
```

## Connection issues

### ACARSHub connectivity problems

#### Symptoms
- Application starts but no messages are processed
- Log messages: "Failed to connect to ACARSHub"
- Connection timeout errors

#### Diagnostic steps

1. **Verify ACARSHub availability**:
   ```bash
   # Test ACARS port connectivity
   telnet $ACARSHUB_HOST 15550
   nc -zv $ACARSHUB_HOST 15550
   
   # Test VDLM2 port connectivity  
   telnet $ACARSHUB_HOST 15555
   nc -zv $ACARSHUB_HOST 15555
   ```

2. **Check network configuration**:
   ```bash
   # DNS resolution
   nslookup $ACARSHUB_HOST
   dig $ACARSHUB_HOST
   
   # Route verification
   traceroute $ACARSHUB_HOST
   ping $ACARSHUB_HOST
   ```

3. **Firewall and security**:
   ```bash
   # Check local firewall rules
   sudo iptables -L | grep 15550
   sudo iptables -L | grep 15555
   
   # SELinux status (if applicable)
   getenforce
   sudo sealert -a /var/log/audit/audit.log
   ```

#### Resolution strategies

**Configuration fixes**:
```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: correct-acarshub-hostname  # Verify hostname
      Port: 15550                      # Verify port number
    VDLM2:
      Host: correct-acarshub-hostname
      Port: 15555
    MaxConcurrentRequests: 5           # Reduce load on ACARSHub
```

**Network troubleshooting**:
```bash
# Test with different network interface
ip route show
ip addr show

# Check for competing processes
netstat -tlnp | grep :15550
lsof -i :15550
```

**Docker-specific issues**:
```bash
# Check container networking
docker network ls
docker network inspect bridge

# Container connectivity test
docker exec acars-processor ping acarshub
docker exec acars-processor telnet acarshub 15550
```

### External API connectivity

#### ADS-B Exchange API issues

**Symptoms**:
- "API rate limit exceeded" warnings
- "Invalid API key" errors
- Timeout failures

**Diagnostic procedures**:
```bash
# Test API connectivity
curl -H "X-API-Key: $ADSB_API_KEY" \
     "https://adsbexchange.com/api/aircraft/json/" \
     -v --connect-timeout 10

# Check rate limiting
curl -I -H "X-API-Key: $ADSB_API_KEY" \
     "https://adsbexchange.com/api/aircraft/json/"
```

**Resolution strategies**:
```yaml
Annotators:
  ADSBExchange:
    APIKey: "${ADSB_API_KEY}"          # Verify API key validity
    # Add retry configuration
    MaxRetryAttempts: 3
    RetryDelaySeconds: 10
```

#### OpenAI API connectivity

**Common issues**:
- API key authentication failures
- Rate limiting and quota exceeded
- Model unavailability

**Diagnostic commands**:
```bash
# Test API authentication
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "test"}], "max_tokens": 5}' \
     https://api.openai.com/v1/chat/completions

# Check account status
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/usage
```

**Configuration adjustments**:
```yaml
Filters:
  OpenAI:
    Model: "gpt-3.5-turbo"             # Use different model if gpt-4 unavailable
    Timeout: 30                        # Increase timeout for reliability
    MaxRetryAttempts: 3
```

#### Ollama connectivity issues

**Local Ollama server problems**:
```bash
# Check Ollama service status
curl http://localhost:11434/api/tags

# Verify model availability
curl http://localhost:11434/api/show -d '{"name": "gemma2:9b"}'

# Check Ollama logs
docker logs ollama
journalctl -u ollama
```

**Resolution steps**:
```bash
# Pull required model
ollama pull gemma2:9b

# Restart Ollama service
docker restart ollama
sudo systemctl restart ollama
```

## Performance issues

### High CPU usage

#### Symptoms
- System responsiveness degradation
- High load averages
- Slow message processing

#### Diagnostic analysis

**CPU profiling**:
```bash
# Monitor CPU usage by process
top -p $(pgrep acars-processor)
htop -p $(pgrep acars-processor)

# Detailed process analysis
strace -p $(pgrep acars-processor) -c -f

# Docker container resource usage
docker stats acars-processor
```

**Configuration optimization**:
```yaml
ACARSProcessorSettings:
  ACARSHub:
    MaxConcurrentRequests: 10          # Reduce concurrent processing
    
Filters:
  OpenAI:
    Timeout: 10                        # Reduce API wait time
  Ollama:
    Timeout: 5                         # Reduce local AI processing time
```

### High memory usage

#### Memory leak detection

**Memory monitoring**:
```bash
# Memory usage tracking
ps -o pid,vsz,rss,comm $(pgrep acars-processor)
pmap $(pgrep acars-processor)

# Container memory usage
docker stats acars-processor --format "table {{.MemUsage}}\t{{.MemPerc}}"
```

**Ollama-related memory issues**:
```yaml
Annotators:
  Ollama:
    Model: "gemma2:2b"                 # Use smaller model
    # Consider disabling if memory constrained
    Enabled: false
```

**Database optimization**:
```yaml
Database:
  Type: sqlite                         # Use SQLite for lower memory usage
  # Or optimize MySQL connection pooling
  ConnectionString: "user:pass@tcp(host:3306)/db?maxOpenConns=10&maxIdleConns=2"
```

### Message processing delays

#### Queue backlog analysis

**Diagnostic commands**:
```bash
# Monitor message queue depth
grep "queue.*backlog" /var/log/acars-processor.log
grep "Processing message" /var/log/acars-processor.log | tail -20

# Check processing latency
grep -E "(Processing|sending)" /var/log/acars-processor.log | tail -20
```

**Performance tuning**:
```yaml
ACARSProcessorSettings:
  ACARSHub:
    MaxConcurrentRequests: 20          # Increase throughput

Filters:
  # Disable expensive filters temporarily
  OpenAI:
    Enabled: false
  Ollama:
    Enabled: false
```

## Configuration issues

### Environment variable problems

#### Missing or incorrect environment variables

**Diagnostic steps**:
```bash
# Check environment variable values
echo $ACARSHUB_HOST
echo $DISCORD_WEBHOOK_URL
echo $ADSB_API_KEY

# Verify variable substitution
grep -r "\${" config.yaml
```

**Common issues and fixes**:
```bash
# Unquoted values with special characters
# Wrong:
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/123/abc?wait=true

# Correct:
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/123/abc?wait=true"
```

**Docker environment variable issues**:
```bash
# Check container environment
docker exec acars-processor env | grep -E "(DISCORD|ADSB|OPENAI)"

# Debug Docker Compose environment
docker-compose config
```

### Schema validation failures

#### Configuration structure errors

**JSON schema validation**:
```bash
# Install validator
npm install -g ajv-cli

# Validate configuration
./acars-processor -s  # Generate current schema
ajv validate -s schema.json -d config.yaml
```

**Common schema violations**:
```yaml
# Missing required fields
ACARSProcessorSettings:
  ACARSHub:
    # Missing ACARS and VDLM2 sections
    MaxConcurrentRequests: 10

# Incorrect data types
Filters:
  Generic:
    AboveSignaldBm: "-50.0"            # Wrong: string instead of number
    AboveSignaldBm: -50.0              # Correct: number
```

## Database issues

### SQLite problems

#### Database lock issues

**Symptoms**:
- "Database locked" errors
- Application hangs during startup
- Failed message saves

**Resolution steps**:
```bash
# Check for multiple processes accessing database
lsof ./messages.db
fuser -v ./messages.db

# Fix database lock
sqlite3 ./messages.db ".timeout 30000"

# Check database integrity
sqlite3 ./messages.db "PRAGMA integrity_check;"
```

**Prevention configuration**:
```yaml
Database:
  Type: sqlite
  SQLiteDatabasePath: ./messages.db
  # Ensure only one acars-processor instance per database file
```

#### File permission issues

**Diagnostic commands**:
```bash
# Check file permissions
ls -la ./messages.db
ls -la ./        # Check directory permissions

# Fix permissions
chmod 644 ./messages.db
chown acars:acars ./messages.db
```

### MySQL connectivity issues

#### Connection string problems

**Common connection string issues**:
```yaml
Database:
  # Wrong: Missing required parameters
  ConnectionString: "user:pass@tcp(host:3306)/db"
  
  # Correct: Include charset and time parsing
  ConnectionString: "user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"
```

**Connection testing**:
```bash
# Test MySQL connectivity
mysql -h $DB_HOST -u $DB_USER -p$DB_PASS $DB_NAME -e "SELECT 1;"

# Check connection limits
mysql -h $DB_HOST -u $DB_USER -p$DB_PASS -e "SHOW PROCESSLIST;"
mysql -h $DB_HOST -u $DB_USER -p$DB_PASS -e "SHOW VARIABLES LIKE 'max_connections';"
```

## Docker-specific issues

### Container startup failures

#### Image and build problems

**Build issues**:
```bash
# Rebuild with verbose output
docker build --no-cache --progress=plain -t acars-processor .

# Check build logs
docker build -t acars-processor . 2>&1 | tee build.log
```

**Container startup debugging**:
```bash
# Run container interactively
docker run -it acars-processor /bin/sh

# Check container logs
docker logs acars-processor --timestamps
docker logs acars-processor --follow
```

### Docker Compose issues

#### Service dependency problems

**Common Docker Compose issues**:
```yaml
# Wrong: Missing dependency
services:
  acars-processor:
    image: acars-processor
    # Missing depends_on

# Correct: Proper dependencies
services:
  acars-processor:
    image: acars-processor
    depends_on:
      - acarshub
      - mysql
```

**Debugging Docker Compose**:
```bash
# Validate Docker Compose configuration
docker-compose config

# Check service status
docker-compose ps
docker-compose logs acars-processor

# Restart specific service
docker-compose restart acars-processor
```

## Receiver-specific issues

### Discord webhook failures

#### Webhook URL problems

**Diagnostic tests**:
```bash
# Test webhook URL manually
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"content": "Test message"}' \
  "$DISCORD_WEBHOOK_URL"
```

**Common issues**:
```yaml
Receivers:
  DiscordWebhook:
    URL: "${DISCORD_WEBHOOK_URL}"      # Ensure proper environment variable
    RequiredFields:                    # Check required fields availability
      - acarsMessageText
      - acarsFlightNumber
```

### New Relic integration issues

**API key validation**:
```bash
# Test New Relic API
curl -X POST \
  -H "Api-Key: $NEWRELIC_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"eventType": "test", "message": "test"}' \
  https://insights-collector.newrelic.com/v1/accounts/$ACCOUNT_ID/events
```

## Recovery procedures

### Service recovery

**Graceful restart procedure**:
```bash
# Stop service gracefully
kill -TERM $(pgrep acars-processor)

# Wait for graceful shutdown
sleep 10

# Force kill if necessary
kill -KILL $(pgrep acars-processor)

# Restart service
./acars-processor -c config.yaml &
```

**Docker restart procedure**:
```bash
# Graceful restart
docker-compose restart acars-processor

# Complete rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Data recovery

**Database recovery**:
```bash
# SQLite recovery
sqlite3 ./messages.db ".recover" | sqlite3 ./messages_recovered.db

# MySQL recovery
mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASS $DB_NAME > backup.sql
mysql -h $DB_HOST -u $DB_USER -p$DB_PASS $DB_NAME < backup.sql
```

This troubleshooting guide provides systematic approaches to diagnosing and resolving common issues across all aspects of ACARS processor deployment and operation.
