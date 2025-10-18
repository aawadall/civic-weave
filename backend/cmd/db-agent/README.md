# Database Agent Cloud Deployment Guide

## Overview

This guide covers deploying the gRPC Database Agent to cloud environments with proper security, scalability, and monitoring.

## Architecture

```
┌─────────────────┐    gRPC/TLS     ┌─────────────────┐    PostgreSQL    ┌─────────────────┐
│   Local Client  │ ──────────────► │  Agent Server   │ ───────────────► │ Metadata DB     │
│   (CLI Tool)    │                 │   (Cloud)       │                  │   (Tracking)    │
└─────────────────┘                 └─────────────────┘                  └─────────────────┘
                                            │
                                            │ PostgreSQL
                                            ▼
                                    ┌─────────────────┐
                                    │ Target Databases│
                                    │ (Production)    │
                                    └─────────────────┘
```

## Prerequisites

1. **Cloud Provider Account** (AWS, GCP, Azure)
2. **PostgreSQL Database** for metadata storage
3. **Target Databases** to manage
4. **SSL Certificates** for TLS encryption
5. **API Keys** generated using `db-keygen`

## Quick Start

### 1. Generate API Keys

```bash
# Generate server keys (run once on agent machine)
cd backend
go run cmd/db-keygen/main.go -server -description="Production Agent Keys"

# Generate client keys (for each developer/CI system)
go run cmd/db-keygen/main.go -client -agent-url=your-agent-host:50051 -description="Developer Keys"
```

### 2. Configure Environment

Create `.env` file:

```bash
# Agent Configuration
AGENT_HOST=0.0.0.0
AGENT_PORT=50051
AGENT_TLS=true
AGENT_CERT_FILE=/etc/ssl/certs/agent.crt
AGENT_KEY_FILE=/etc/ssl/private/agent.key

# Metadata Database
METADB_HOST=your-metadata-db-host
METADB_PORT=5432
METADB_USER=db_agent_user
METADB_PASSWORD=secure_password
METADB_NAME=db_agent_metadata
METADB_SSL_MODE=require

# Security
ENABLE_AUTH=true
ENABLE_RATE_LIMIT=true
RATE_LIMIT_RPS=100
LOG_LEVEL=info
```

### 3. Run Agent

```bash
# Development
cd backend
go run cmd/db-agent/main.go -port=50051 -tls=false

# Production with TLS
go run cmd/db-agent/main.go -port=50051 -tls -cert=/path/to/cert.crt -key=/path/to/key.key
```

## Cloud Deployment Options

### Option 1: Docker Deployment

#### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o db-agent ./cmd/db-agent

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/db-agent .
COPY --from=builder /app/keys/server/ ./keys/server/

EXPOSE 50051

CMD ["./db-agent", "-port=50051", "-tls", "-cert=./keys/server/server.crt", "-key=./keys/server/server.key"]
```

#### Docker Compose

```yaml
version: '3.8'
services:
  db-agent:
    build: .
    ports:
      - "50051:50051"
    environment:
      - METADB_HOST=metadata-db
      - METADB_USER=db_agent_user
      - METADB_PASSWORD=secure_password
      - METADB_NAME=db_agent_metadata
    depends_on:
      - metadata-db
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "./db-agent", "-command=ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  metadata-db:
    image: postgres:15
    environment:
      POSTGRES_DB: db_agent_metadata
      POSTGRES_USER: db_agent_user
      POSTGRES_PASSWORD: secure_password
    volumes:
      - metadata_data:/var/lib/postgresql/data
      - ./backend/pkg/metadb/schema.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5433:5432"

volumes:
  metadata_data:
```

### Option 2: Kubernetes Deployment

#### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: db-agent
```

#### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: db-agent-config
  namespace: db-agent
data:
  METADB_HOST: "metadata-db.db-agent.svc.cluster.local"
  METADB_PORT: "5432"
  METADB_NAME: "db_agent_metadata"
  METADB_SSL_MODE: "require"
  ENABLE_AUTH: "true"
  ENABLE_RATE_LIMIT: "true"
  RATE_LIMIT_RPS: "100"
  LOG_LEVEL: "info"
```

#### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-agent-secrets
  namespace: db-agent
type: Opaque
data:
  METADB_PASSWORD: <base64-encoded-password>
  METADB_USER: <base64-encoded-username>
```

#### TLS Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-agent-tls
  namespace: db-agent
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-certificate>
  tls.key: <base64-encoded-private-key>
```

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: db-agent
  namespace: db-agent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: db-agent
  template:
    metadata:
      labels:
        app: db-agent
    spec:
      containers:
      - name: db-agent
        image: your-registry/db-agent:latest
        ports:
        - containerPort: 50051
        envFrom:
        - configMapRef:
            name: db-agent-config
        - secretRef:
            name: db-agent-secrets
        volumeMounts:
        - name: tls-certs
          mountPath: /etc/ssl/certs
          readOnly: true
        - name: tls-keys
          mountPath: /etc/ssl/private
          readOnly: true
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - "grpc_health_probe -addr=localhost:50051"
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - "grpc_health_probe -addr=localhost:50051"
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: tls-certs
        secret:
          secretName: db-agent-tls
          items:
          - key: tls.crt
            path: agent.crt
      - name: tls-keys
        secret:
          secretName: db-agent-tls
          items:
          - key: tls.key
            path: agent.key
```

#### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: db-agent-service
  namespace: db-agent
spec:
  selector:
    app: db-agent
  ports:
  - port: 50051
    targetPort: 50051
    protocol: TCP
  type: ClusterIP
```

#### Ingress (with TLS termination)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: db-agent-ingress
  namespace: db-agent
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/backend-protocol: GRPC
spec:
  tls:
  - hosts:
    - db-agent.yourdomain.com
    secretName: db-agent-tls-ingress
  rules:
  - host: db-agent.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: db-agent-service
            port:
              number: 50051
```

### Option 3: AWS ECS/Fargate

#### Task Definition

```json
{
  "family": "db-agent",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/db-agent-task-role",
  "containerDefinitions": [
    {
      "name": "db-agent",
      "image": "your-account.dkr.ecr.region.amazonaws.com/db-agent:latest",
      "portMappings": [
        {
          "containerPort": 50051,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "METADB_HOST",
          "value": "your-rds-endpoint.region.rds.amazonaws.com"
        },
        {
          "name": "METADB_PORT",
          "value": "5432"
        },
        {
          "name": "METADB_NAME",
          "value": "db_agent_metadata"
        },
        {
          "name": "ENABLE_AUTH",
          "value": "true"
        }
      ],
      "secrets": [
        {
          "name": "METADB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:db-agent/metadb-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/db-agent",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "grpc_health_probe -addr=localhost:50051"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

### Option 4: Google Cloud Run

#### Cloud Run Service

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: db-agent
  namespace: default
  annotations:
    run.googleapis.com/ingress: all
    run.googleapis.com/execution-environment: gen2
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "10"
        autoscaling.knative.dev/minScale: "1"
        run.googleapis.com/cpu-throttling: "false"
        run.googleapis.com/execution-environment: gen2
    spec:
      containerConcurrency: 100
      timeoutSeconds: 300
      containers:
      - image: gcr.io/your-project/db-agent:latest
        ports:
        - containerPort: 50051
        env:
        - name: METADB_HOST
          value: "your-cloud-sql-ip"
        - name: METADB_PORT
          value: "5432"
        - name: METADB_NAME
          value: "db_agent_metadata"
        - name: METADB_USER
          valueFrom:
            secretKeyRef:
              name: db-agent-secrets
              key: username
        - name: METADB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-agent-secrets
              key: password
        resources:
          limits:
            cpu: "2"
            memory: "2Gi"
          requests:
            cpu: "1"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /
            port: 50051
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 50051
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Security Configuration

### 1. TLS Certificates

#### Self-Signed (Development)

```bash
# Generate private key
openssl genrsa -out server.key 2048

# Generate certificate
openssl req -new -x509 -key server.key -out server.crt -days 365 \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=your-agent-host"
```

#### Let's Encrypt (Production)

```bash
# Using certbot
certbot certonly --standalone -d your-agent-host.com

# Certificates will be in /etc/letsencrypt/live/your-agent-host.com/
```

### 2. Firewall Rules

#### AWS Security Groups

```bash
# Allow gRPC traffic from specific IPs
aws ec2 authorize-security-group-ingress \
  --group-id sg-12345678 \
  --protocol tcp \
  --port 50051 \
  --cidr 10.0.0.0/8
```

#### GCP Firewall

```bash
gcloud compute firewall-rules create allow-db-agent \
  --allow tcp:50051 \
  --source-ranges 10.0.0.0/8 \
  --target-tags db-agent
```

### 3. Network Security

- Use VPC/Private Networks
- Enable TLS encryption
- Restrict access to metadata database
- Use API key authentication
- Implement rate limiting
- Enable audit logging

## Monitoring and Logging

### 1. Health Checks

```bash
# Using grpc_health_probe
grpc_health_probe -addr=your-agent-host:50051

# Using client ping
go run cmd/db-client/main.go -command=ping -agent=your-agent-host:50051
```

### 2. Metrics Collection

Add Prometheus metrics to the agent:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_agent_requests_total",
            Help: "Total number of requests",
        },
        []string{"method", "status"},
    )
    
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_agent_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"method"},
    )
)
```

### 3. Log Aggregation

#### Docker Logging

```yaml
# docker-compose.yml
services:
  db-agent:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

#### Kubernetes Logging

```yaml
# Use Fluentd or similar for log aggregation
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/db-agent*.log
      pos_file /var/log/fluentd-db-agent.log.pos
      tag kubernetes.db-agent
      format json
    </source>
```

## Database Setup

### 1. Metadata Database

```sql
-- Create metadata database
CREATE DATABASE db_agent_metadata;

-- Create user
CREATE USER db_agent_user WITH PASSWORD 'secure_password';

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE db_agent_metadata TO db_agent_user;

-- Connect to metadata database and run schema
\c db_agent_metadata
\i backend/pkg/metadb/schema.sql
```

### 2. Target Database Registration

```bash
# Register a database for management
go run cmd/db-client/main.go -command=bootstrap \
  -manifest=./manifest \
  -database=production \
  -agent=your-agent-host:50051
```

## Backup and Recovery

### 1. Metadata Database Backup

```bash
# Daily backup
pg_dump -h metadata-db-host -U db_agent_user db_agent_metadata > backup_$(date +%Y%m%d).sql

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h $METADB_HOST -U $METADB_USER $METADB_NAME > $BACKUP_DIR/metadata_$DATE.sql
find $BACKUP_DIR -name "metadata_*.sql" -mtime +30 -delete
```

### 2. Agent Configuration Backup

```bash
# Backup keys and configuration
tar -czf agent-config-backup.tar.gz keys/ .env
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   ```bash
   # Check if agent is running
   netstat -tlnp | grep 50051
   
   # Check firewall
   iptables -L | grep 50051
   ```

2. **TLS Certificate Issues**
   ```bash
   # Verify certificate
   openssl x509 -in server.crt -text -noout
   
   # Test TLS connection
   openssl s_client -connect your-agent-host:50051
   ```

3. **Database Connection Issues**
   ```bash
   # Test metadata database connection
   psql -h $METADB_HOST -U $METADB_USER -d $METADB_NAME -c "SELECT 1;"
   ```

### Debug Mode

```bash
# Run agent with debug logging
go run cmd/db-agent/main.go -port=50051 -log-level=debug

# Test client connection
go run cmd/db-client/main.go -command=ping -agent=localhost:50051 -verbose
```

## Performance Tuning

### 1. Connection Pooling

```go
// In agent configuration
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 2. Resource Limits

```yaml
# Kubernetes resource limits
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### 3. Scaling

- Use multiple agent instances behind load balancer
- Implement horizontal pod autoscaling
- Use connection pooling for target databases
- Cache frequently accessed metadata

## Maintenance

### 1. Updates

```bash
# Update agent
docker pull your-registry/db-agent:latest
docker-compose up -d db-agent

# Update client
go get -u ./cmd/db-client
go build ./cmd/db-client
```

### 2. Key Rotation

```bash
# Generate new keys
go run cmd/db-keygen/main.go -server -description="New Production Keys"

# Update agent with new keys
# Update all clients with new keys
```

### 3. Monitoring

- Monitor agent health and performance
- Track deployment success rates
- Monitor database connection health
- Alert on authentication failures

This guide provides comprehensive deployment options for the gRPC Database Agent. Choose the deployment method that best fits your infrastructure and requirements.
