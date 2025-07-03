# Deployment

## Docker deployment

Create configuration:
```bash
curl -o config.yaml https://raw.githubusercontent.com/tyzbit/acars-processor/main/config_example.yaml
# Edit configuration as needed
```

Build and run:
```bash
docker build -t acars-processor .
docker run -d --name acars-processor \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/data:/app/data \
  acars-processor
```

## Binary deployment

Download and run:
```bash
wget https://github.com/tyzbit/acars-processor/releases/latest/download/acars-processor-linux-amd64
chmod +x acars-processor-linux-amd64
wget https://raw.githubusercontent.com/tyzbit/acars-processor/main/config_example.yaml -O config.yaml
# Edit configuration as needed
./acars-processor-linux-amd64 -c config.yaml
```

## Configuration examples

### Development setup
```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: localhost
      Port: 15550
  Database:
    Type: sqlite
    SQLiteDatabasePath: ./messages.db
  LogLevel: debug

Annotators:
  ACARS:
    Enabled: true

Filters:
  Generic:
    HasText: true

Receivers:
  DiscordWebhook:
    URL: "${DISCORD_WEBHOOK_URL}"
```

### Production setup
```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: "${ACARSHUB_HOST}"
      Port: 15550
  Database:
    Type: mariadb
    ConnectionString: "${DB_CONNECTION_STRING}"
  LogLevel: info

Annotators:
  ACARS:
    Enabled: true
  ADSBExchange:
    Enabled: true
    APIKey: "${ADSB_API_KEY}"

Filters:
  Generic:
    HasText: true
    Emergency: true

Receivers:
  DiscordWebhook:
    URL: "${DISCORD_WEBHOOK_URL}"
  NewRelic:
    Enabled: true
    APIKey: "${NEWRELIC_API_KEY}"
```

## Database setup

### SQLite (development)
Default configuration creates SQLite database automatically.

### MariaDB (production)
Create database and user:
```sql
CREATE DATABASE acars_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'acars'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON acars_db.* TO 'acars'@'%';
FLUSH PRIVILEGES;
```

Configure connection:
```yaml
ACARSProcessorSettings:
  Database:
    Type: mariadb
    ConnectionString: "user:password@tcp(host:3306)/acars_db?charset=utf8mb4&parseTime=True&loc=Local"
```
