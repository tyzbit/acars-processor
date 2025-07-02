# Deployment guide

## Overview

This guide provides comprehensive deployment instructions for ACARS processor across different environments, from development to production. The application supports various deployment methods including Docker containers, Kubernetes, and direct installation.

## Prerequisites

### System requirements

**Minimum requirements**:
- CPU: 1 core, 2.0 GHz
- RAM: 512 MB (8 GB recommended with Ollama)
- Storage: 1 GB (50 GB recommended with Ollama models)
- Network: Outbound HTTPS access for external APIs

**Recommended production requirements**:
- CPU: 4 cores, 2.4 GHz
- RAM: 8 GB (16 GB with multiple Ollama models)
- Storage: 100 GB SSD
- Network: Dedicated network segment with monitoring

### Infrastructure dependencies

**Required services**:
- ACARSHub instance for message streams
- Database (SQLite for development, MySQL/PostgreSQL for production)

**Optional services**:
- Ollama server for local AI processing
- External monitoring (New Relic, Datadog, etc.)
- Log aggregation (ELK stack, Grafana Loki, etc.)

## Environment preparation

### Docker environment

1. **Install Docker and Docker Compose**:
   ```bash
   # Ubuntu/Debian
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo usermod -aG docker $USER
   
   # Install Docker Compose
   sudo curl -L "https://github.com/docker/compose/releases/download/v2.21.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```

2. **Verify installation**:
   ```bash
   docker --version
   docker-compose --version
   ```

### Kubernetes environment

1. **Install kubectl**:
   ```bash
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
   ```

2. **Verify cluster access**:
   ```bash
   kubectl cluster-info
   kubectl get nodes
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

### Environment variables

Create environment-specific variable files:

**Development (.env.dev)**:
```bash
# ACARSHub connection
ACARSHUB_HOST=localhost
ACARS_PORT=15550
VDLM2_PORT=15555

# Database
DATABASE_TYPE=sqlite
SQLITE_PATH=./dev_messages.db

# Optional services
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/dev/token
OLLAMA_URL=http://localhost:11434

# Logging
LOG_LEVEL=debug
COLOR_OUTPUT=true
```

**Production (.env.prod)**:
```bash
# ACARSHub connection
ACARSHUB_HOST=acarshub.internal
ACARS_PORT=15550
VDLM2_PORT=15555

# Database
DATABASE_TYPE=mysql
DATABASE_CONNECTION_STRING=acars:${DB_PASSWORD}@tcp(mysql:3306)/acars_db?charset=utf8mb4&parseTime=True&loc=Local

# External services
DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL}
OPENAI_API_KEY=${OPENAI_API_KEY}
ADSB_API_KEY=${ADSB_API_KEY}
NEWRELIC_API_KEY=${NEWRELIC_API_KEY}

# Performance
MAX_CONCURRENT_REQUESTS=20
LOG_LEVEL=info
COLOR_OUTPUT=false
```

### Configuration templates

**Base configuration (config.base.yaml)**:
```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: "${ACARSHUB_HOST}"
      Port: ${ACARS_PORT}
    VDLM2:
      Host: "${ACARSHUB_HOST}"
      Port: ${VDLM2_PORT}
    MaxConcurrentRequests: ${MAX_CONCURRENT_REQUESTS}
  Database:
    Type: "${DATABASE_TYPE}"
    SQLiteDatabasePath: "${SQLITE_PATH}"
    ConnectionString: "${DATABASE_CONNECTION_STRING}"
  LogLevel: "${LOG_LEVEL}"
  ColorOutput: ${COLOR_OUTPUT}

Annotators:
  ACARS:
    Enabled: true
  ADSBExchange:
    Enabled: false
    APIKey: "${ADSB_API_KEY}"
  Ollama:
    Enabled: false
    URL: "${OLLAMA_URL}"
    Model: "gemma2:9b"

Filters:
  Generic:
    HasText: true
    Emergency: true

Receivers:
  DiscordWebhook:
    Enabled: true
    URL: "${DISCORD_WEBHOOK_URL}"
```

## Deployment methods

### Method 1: Docker Compose (recommended for most use cases)

1. **Create deployment directory**:
   ```bash
   mkdir -p /opt/acars-processor/{config,data,logs}
   cd /opt/acars-processor
   ```

2. **Create docker-compose.yml**:
   ```yaml
   version: '3.8'
   
   services:
     acars-processor:
       image: ghcr.io/tyzbit/acars-processor:latest
       container_name: acars-processor
       restart: unless-stopped
       volumes:
         - ./config:/app/config:ro
         - ./data:/app/data
         - ./logs:/app/logs
       environment:
         - CONFIG_PATH=/app/config/config.yaml
       env_file:
         - .env.prod
       depends_on:
         - mysql
         - acarshub
       networks:
         - acars-network
   
     mysql:
       image: mysql:8.0
       container_name: acars-mysql
       restart: unless-stopped
       environment:
         - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
         - MYSQL_DATABASE=acars_db
         - MYSQL_USER=acars
         - MYSQL_PASSWORD=${DB_PASSWORD}
       volumes:
         - mysql-data:/var/lib/mysql
         - ./mysql/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
       networks:
         - acars-network
   
     acarshub:
       image: fredclausen/acarshub:latest
       container_name: acarshub
       restart: unless-stopped
       ports:
         - "15550:15550"
         - "15555:15555"
         - "80:80"
       environment:
         - ACARS_LISTEN_PORT=15550
         - VDLM2_LISTEN_PORT=15555
       networks:
         - acars-network
   
     ollama:
       image: ollama/ollama:latest
       container_name: ollama
       restart: unless-stopped
       volumes:
         - ollama-data:/root/.ollama
       environment:
         - OLLAMA_HOST=0.0.0.0
       networks:
         - acars-network
       deploy:
         resources:
           reservations:
             devices:
               - driver: nvidia
                 count: 1
                 capabilities: [gpu]
   
   volumes:
     mysql-data:
     ollama-data:
   
   networks:
     acars-network:
       driver: bridge
   ```

3. **Deploy the stack**:
   ```bash
   # Load environment variables
   set -a && source .env.prod && set +a
   
   # Start services
   docker-compose up -d
   
   # Verify deployment
   docker-compose logs -f acars-processor
   ```

### Method 2: Kubernetes deployment

1. **Create namespace**:
   ```bash
   kubectl create namespace acars-processor
   ```

2. **Create configuration ConfigMap**:
   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: acars-config
     namespace: acars-processor
   data:
     config.yaml: |
       ACARSProcessorSettings:
         ACARSHub:
           ACARS:
             Host: "acarshub"
             Port: 15550
           VDLM2:
             Host: "acarshub"
             Port: 15555
           MaxConcurrentRequests: 20
         Database:
           Type: "mysql"
           ConnectionString: "acars:password@tcp(mysql:3306)/acars_db?charset=utf8mb4&parseTime=True&loc=Local"
         LogLevel: "info"
         ColorOutput: false
       # ... rest of configuration
   ```

3. **Create secrets**:
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: acars-secrets
     namespace: acars-processor
   type: Opaque
   data:
     discord-webhook-url: <base64-encoded-webhook-url>
     openai-api-key: <base64-encoded-api-key>
     adsb-api-key: <base64-encoded-api-key>
     mysql-password: <base64-encoded-password>
   ```

4. **Create deployment**:
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: acars-processor
     namespace: acars-processor
   spec:
     replicas: 2
     selector:
       matchLabels:
         app: acars-processor
     template:
       metadata:
         labels:
           app: acars-processor
       spec:
         containers:
         - name: acars-processor
           image: ghcr.io/tyzbit/acars-processor:latest
           args: ["-c", "/config/config.yaml"]
           env:
           - name: DISCORD_WEBHOOK_URL
             valueFrom:
               secretKeyRef:
                 name: acars-secrets
                 key: discord-webhook-url
           - name: OPENAI_API_KEY
             valueFrom:
               secretKeyRef:
                 name: acars-secrets
                 key: openai-api-key
           volumeMounts:
           - name: config
             mountPath: /config
             readOnly: true
           resources:
             requests:
               memory: "512Mi"
               cpu: "500m"
             limits:
               memory: "2Gi"
               cpu: "2000m"
           livenessProbe:
             exec:
               command:
               - /bin/sh
               - -c
               - "pgrep -f acars-processor"
             initialDelaySeconds: 30
             periodSeconds: 30
           readinessProbe:
             exec:
               command:
               - /bin/sh
               - -c
               - "pgrep -f acars-processor"
             initialDelaySeconds: 10
             periodSeconds: 10
         volumes:
         - name: config
           configMap:
             name: acars-config
   ```

5. **Deploy to Kubernetes**:
   ```bash
   kubectl apply -f k8s/
   kubectl get pods -n acars-processor
   kubectl logs -f deployment/acars-processor -n acars-processor
   ```

### Method 3: Direct installation

1. **Create service user**:
   ```bash
   sudo useradd -r -s /bin/false acars
   sudo mkdir -p /opt/acars-processor/{bin,config,data,logs}
   sudo chown -R acars:acars /opt/acars-processor
   ```

2. **Build and install**:
   ```bash
   git clone https://github.com/tyzbit/acars-processor.git
   cd acars-processor
   go build -ldflags="-s -w" -o acars-processor
   sudo cp acars-processor /opt/acars-processor/bin/
   sudo cp config_example.yaml /opt/acars-processor/config/config.yaml
   ```

3. **Create systemd service**:
   ```ini
   [Unit]
   Description=ACARS Processor
   After=network.target
   
   [Service]
   Type=simple
   User=acars
   Group=acars
   WorkingDirectory=/opt/acars-processor
   ExecStart=/opt/acars-processor/bin/acars-processor -c /opt/acars-processor/config/config.yaml
   Restart=always
   RestartSec=10
   StandardOutput=journal
   StandardError=journal
   
   # Security settings
   NoNewPrivileges=true
   PrivateTmp=true
   ProtectSystem=strict
   ProtectHome=true
   ReadWritePaths=/opt/acars-processor/data /opt/acars-processor/logs
   
   [Install]
   WantedBy=multi-user.target
   ```

4. **Enable and start service**:
   ```bash
   sudo systemctl enable acars-processor
   sudo systemctl start acars-processor
   sudo systemctl status acars-processor
   ```

## Database setup

### SQLite (development/single instance)

SQLite requires no additional setup - the database file is created automatically.

**Configuration**:
```yaml
Database:
  Type: sqlite
  SQLiteDatabasePath: ./messages.db
```

### MySQL/MariaDB (production)

1. **Install MySQL**:
   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install mysql-server
   sudo mysql_secure_installation
   ```

2. **Create database and user**:
   ```sql
   CREATE DATABASE acars_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   CREATE USER 'acars'@'%' IDENTIFIED BY 'secure_password_here';
   GRANT ALL PRIVILEGES ON acars_db.* TO 'acars'@'%';
   FLUSH PRIVILEGES;
   ```

3. **Optimize MySQL configuration** (`/etc/mysql/conf.d/acars.cnf`):
   ```ini
   [mysqld]
   # Performance tuning for ACARS processor
   innodb_buffer_pool_size = 1G
   innodb_log_file_size = 256M
   max_connections = 200
   
   # Character set
   character_set_server = utf8mb4
   collation_server = utf8mb4_unicode_ci
   
   # Binary logging (for replication)
   log-bin = mysql-bin
   server-id = 1
   ```

### PostgreSQL (alternative production option)

1. **Install PostgreSQL**:
   ```bash
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   ```

2. **Create database and user**:
   ```sql
   CREATE DATABASE acars_db WITH ENCODING 'UTF8';
   CREATE USER acars WITH PASSWORD 'secure_password_here';
   GRANT ALL PRIVILEGES ON DATABASE acars_db TO acars;
   ```

## High availability deployment

### Database clustering

**MySQL with replication**:

1. **Primary server configuration** (`/etc/mysql/conf.d/primary.cnf`):
   ```ini
   [mysqld]
   server-id = 1
   log-bin = mysql-bin
   binlog_do_db = acars_db
   ```

2. **Replica server configuration** (`/etc/mysql/conf.d/replica.cnf`):
   ```ini
   [mysqld]
   server-id = 2
   relay-log = mysql-relay-bin
   log_bin = mysql-bin
   read_only = 1
   ```

3. **Setup replication**:
   ```sql
   -- On primary
   CREATE USER 'replication'@'%' IDENTIFIED BY 'replication_password';
   GRANT REPLICATION SLAVE ON *.* TO 'replication'@'%';
   FLUSH PRIVILEGES;
   
   -- On replica
   CHANGE MASTER TO
     MASTER_HOST='primary-ip',
     MASTER_USER='replication',
     MASTER_PASSWORD='replication_password';
   START SLAVE;
   ```

### Load balancing

**HAProxy configuration** (`/etc/haproxy/haproxy.cfg`):
```
global
    daemon
    maxconn 4096

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

frontend acars_frontend
    bind *:80
    default_backend acars_backend

backend acars_backend
    balance roundrobin
    option httpchk GET /health
    server acars1 acars-processor-1:8080 check
    server acars2 acars-processor-2:8080 check
    server acars3 acars-processor-3:8080 check
```

### Container orchestration scaling

**Kubernetes horizontal pod autoscaler**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: acars-processor-hpa
  namespace: acars-processor
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: acars-processor
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## Security hardening

### Network security

1. **Firewall configuration** (UFW):
   ```bash
   # Allow SSH
   sudo ufw allow ssh
   
   # Allow ACARSHub ports (internal only)
   sudo ufw allow from 10.0.0.0/8 to any port 15550
   sudo ufw allow from 10.0.0.0/8 to any port 15555
   
   # Allow database (internal only)
   sudo ufw allow from 10.0.0.0/8 to any port 3306
   
   # Enable firewall
   sudo ufw enable
   ```

2. **TLS/SSL configuration**:
   ```yaml
   # Use TLS for external API calls
   Annotators:
     ADSBExchange:
       APIKey: "${ADSB_API_KEY}"
       # API calls use HTTPS by default
     
   # Use TLS for webhook deliveries
   Receivers:
     Webhook:
       URL: "https://secure-endpoint.example.com/webhook"
   ```

### Container security

1. **Security-focused Dockerfile**:
   ```dockerfile
   FROM golang:1.24-alpine AS builder
   RUN apk add --no-cache git ca-certificates
   WORKDIR /src
   COPY . .
   RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o acars-processor .
   
   FROM scratch
   COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
   COPY --from=builder /src/acars-processor /acars-processor
   USER 65534:65534
   ENTRYPOINT ["/acars-processor"]
   ```

2. **Pod security context**:
   ```yaml
   securityContext:
     runAsNonRoot: true
     runAsUser: 65534
     runAsGroup: 65534
     fsGroup: 65534
     seccompProfile:
       type: RuntimeDefault
     capabilities:
       drop:
       - ALL
   ```

### Secrets management

**HashiCorp Vault integration**:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: acars-processor
  namespace: acars-processor
  annotations:
    vault.hashicorp.com/agent-inject: "true"
    vault.hashicorp.com/agent-inject-secret-config: "secret/acars-processor"
    vault.hashicorp.com/role: "acars-processor"
```

## Monitoring and observability

### Health checks

1. **Application health endpoint** (implement in future version):
   ```go
   http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
       // Check database connectivity
       // Check ACARSHub connectivity
       // Return JSON status
   })
   ```

2. **Docker health check**:
   ```dockerfile
   HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
     CMD pgrep -f acars-processor || exit 1
   ```

### Logging configuration

**Structured logging with centralized collection**:
```yaml
ACARSProcessorSettings:
  LogLevel: info
  ColorOutput: false  # Plain output for log aggregation
```

**Fluentd configuration for log shipping**:
```xml
<source>
  @type forward
  port 24224
  bind 0.0.0.0
</source>

<match acars.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  index_name acars-logs
  type_name _doc
</match>
```

### Log monitoring

**Application logs**:
```bash
# Follow application logs
docker-compose logs -f acars-processor

# View recent errors
docker-compose logs --tail=100 acars-processor | grep -i error

# Check startup messages
docker-compose logs acars-processor | head -50
```

**Log aggregation setup**:
{
  "dashboard": {
    "title": "ACARS Processor Metrics",
    "panels": [
      {
        "title": "Messages Processed/sec",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(acars_messages_processed_total[5m])"
          }
        ]
      }
    ]
  }
}
```

## Backup and recovery

### Database backup

**Automated MySQL backup script**:
```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups/acars"
DB_NAME="acars_db"
DB_USER="backup_user"
DB_PASS="backup_password"

mkdir -p $BACKUP_DIR

# Create backup
mysqldump -u$DB_USER -p$DB_PASS $DB_NAME | gzip > $BACKUP_DIR/acars_backup_$DATE.sql.gz

# Retain only last 30 days
find $BACKUP_DIR -name "acars_backup_*.sql.gz" -mtime +30 -delete

# Verify backup
if [ $? -eq 0 ]; then
    echo "Backup completed successfully: acars_backup_$DATE.sql.gz"
else
    echo "Backup failed!"
    exit 1
fi
```

**Cron schedule for automated backups**:
```bash
# Backup every 6 hours
0 */6 * * * /opt/scripts/backup_acars.sh >> /var/log/acars_backup.log 2>&1
```

### Configuration backup

**Backup configuration and deployment files**:
```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups/config"
CONFIG_DIRS="/opt/acars-processor/config /opt/acars-processor/k8s"

mkdir -p $BACKUP_DIR

tar czf $BACKUP_DIR/acars_config_$DATE.tar.gz $CONFIG_DIRS

# Retain only last 90 days
find $BACKUP_DIR -name "acars_config_*.tar.gz" -mtime +90 -delete
```

### Disaster recovery procedures

1. **Recovery checklist**:
   - [ ] Verify infrastructure availability
   - [ ] Restore database from latest backup
   - [ ] Deploy application from known-good configuration
   - [ ] Verify external service connectivity
   - [ ] Resume message processing
   - [ ] Validate data flow end-to-end

2. **Database recovery**:
   ```bash
   # Stop application
   sudo systemctl stop acars-processor
   
   # Restore database
   gunzip < /opt/backups/acars/acars_backup_latest.sql.gz | mysql -u root -p acars_db
   
   # Start application
   sudo systemctl start acars-processor
   
   # Verify operation
   sudo systemctl status acars-processor
   ```

## Performance tuning

### Application tuning

1. **Concurrency optimization**:
   ```yaml
   ACARSProcessorSettings:
     ACARSHub:
       MaxConcurrentRequests: 20  # Tune based on available CPU cores
   ```

2. **Database connection pooling**:
   ```yaml
   Database:
     ConnectionString: "acars:password@tcp(mysql:3306)/acars_db?charset=utf8mb4&parseTime=True&loc=Local&maxOpenConns=25&maxIdleConns=5"
   ```

### Operating system tuning

1. **Kernel parameters** (`/etc/sysctl.conf`):
   ```ini
   # Increase network buffers
   net.core.rmem_max = 16777216
   net.core.wmem_max = 16777216
   net.ipv4.tcp_rmem = 4096 65536 16777216
   net.ipv4.tcp_wmem = 4096 65536 16777216
   
   # Increase file descriptor limits
   fs.file-max = 100000
   ```

2. **Process limits** (`/etc/security/limits.conf`):
   ```
   acars soft nofile 65536
   acars hard nofile 65536
   acars soft nproc 4096
   acars hard nproc 4096
   ```

### Container resource limits

**Docker Compose resource constraints**:
```yaml
services:
  acars-processor:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

## Troubleshooting

### Common deployment issues

1. **Container startup failures**:
   ```bash
   # Check container logs
   docker logs acars-processor
   
   # Check resource usage
   docker stats acars-processor
   
   # Verify configuration
   docker exec acars-processor cat /app/config/config.yaml
   ```

2. **Database connection issues**:
   ```bash
   # Test database connectivity
   mysql -h mysql -u acars -p acars_db
   
   # Check database logs
   docker logs acars-mysql
   
   # Verify network connectivity
   docker exec acars-processor ping mysql
   ```

3. **External API failures**:
   ```bash
   # Test API connectivity
   curl -H "X-API-Key: $ADSB_API_KEY" "https://adsbexchange.com/api/aircraft/v2/registration/n123ab/"
   
   # Check DNS resolution
   nslookup api.openai.com
   
   # Verify TLS certificates
   openssl s_client -connect api.openai.com:443
   ```

### Performance troubleshooting

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

This deployment guide provides comprehensive instructions for deploying ACARS processor in various environments with appropriate security, monitoring, and performance considerations.
