# Deployment guide

## Overview

This guide provides deployment instructions for ACARS processor. The application supports Docker containers and direct installation.

## Prerequisites

### Infrastructure dependencies

**Required services**:
- ACARSHub instance for message streams
- Database (SQLite for development, MariaDB for production)

**Optional services**:
- Ollama server for local AI processing
- External monitoring (New Relic, etc.)

## Environment preparation

### Docker environment

1. **Install Docker**:
   ```bash
   # Ubuntu/Debian
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo usermod -aG docker $USER
   ```

2. **Verify installation**:
   ```bash
   docker --version
   ```

### Native environment

1. **Install Go 1.24**:
   ```bash
   wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

2. **Verify Go installation**:
   ```bash
   go version
   ```

## Configuration management

Create environment-specific configuration files:

**Development**:
```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: localhost
      Port: 15550
    VDLM2:
      Host: localhost
      Port: 15555
  Database:
    Type: sqlite
    SQLiteDatabasePath: ./dev_messages.db
  LogLevel: debug

Annotators:
  ACARS:
    Enabled: true
  VDLM2:
    Enabled: true

Filters:
  Generic:
    HasText: true

Receivers:
  DiscordWebhook:
    Enabled: true
    URL: "${DISCORD_WEBHOOK_URL}"
```

**Production**:
```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: "${ACARSHUB_HOST}"
      Port: 15550
    VDLM2:
      Host: "${ACARSHUB_HOST}"
      Port: 15555
  Database:
    Type: mariadb
    ConnectionString: "${DB_CONNECTION_STRING}"
  LogLevel: info

Annotators:
  ACARS:
    Enabled: true
  VDLM2:
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
    Enabled: true
    URL: "${DISCORD_WEBHOOK_URL}"
  NewRelic:
    Enabled: true
    APIKey: "${NEWRELIC_API_KEY}"
```

## Deployment methods

### Docker deployment

1. **Build image**:
   ```bash
   docker build -t acars-processor .
   ```

2. **Run container**:
   ```bash
   docker run -d \
     --name acars-processor \
     -v $(pwd)/config.yaml:/app/config.yaml \
     -v $(pwd)/data:/app/data \
     acars-processor
   ```

3. **Monitor logs**:
   ```bash
   docker logs -f acars-processor
   ```

### Direct installation

1. **Clone and build**:
   ```bash
   git clone https://github.com/tyzbit/acars-processor.git
   cd acars-processor
   go mod download
   go build -o acars-processor
   ```

2. **Configure**:
   ```bash
   cp config_example.yaml config.yaml
   nano config.yaml
   ```

3. **Run**:
   ```bash
   ./acars-processor -c config.yaml
   ```

## Database setup

### SQLite (development/single instance)

**Configuration**:
```yaml
ACARSProcessorSettings:
  Database:
    Type: sqlite
    SQLiteDatabasePath: ./messages.db
```

### MariaDB (production)

1. **Install MariaDB**:
   ```bash
   sudo apt update
   sudo apt install mariadb-server mariadb-client
   sudo mysql_secure_installation
   ```

2. **Create database and user**:
   ```sql
   CREATE DATABASE acars_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   CREATE USER 'acars'@'%' IDENTIFIED BY 'secure_password_here';
   GRANT ALL PRIVILEGES ON acars_db.* TO 'acars'@'%';
   FLUSH PRIVILEGES;
   ```

3. **Configuration**:
   ```yaml
   ACARSProcessorSettings:
     Database:
       Type: mariadb
       ConnectionString: "acars:password@tcp(dbhost:3306)/acars_db?charset=utf8mb4&parseTime=True&loc=Local"
   ```

## Performance troubleshooting

1. **High CPU usage**:
   - Monitor Ollama model performance
   - Check concurrent request settings
   - Review message processing rate

2. **High memory usage**:
   - Review Ollama model size
   - Check database connection pooling
   - Monitor message queue depth

3. **Network issues**:
   - Verify ACARSHub connectivity
   - Check external API rate limits
   - Monitor network latency and packet loss
