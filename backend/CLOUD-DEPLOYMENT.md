# Cloud Deployment Guide for Database Agent

## Quick Start

### 1. Prerequisites

- Docker and Docker Compose
- Go 1.21+
- kubectl (for Kubernetes deployment)
- Access to a cloud provider (AWS, GCP, Azure)

### 2. Generate API Keys

```bash
cd backend

# Generate server keys
go run cmd/db-keygen/main.go -server -description="Production Server Keys"

# Generate client keys  
go run cmd/db-keygen/main.go -client -agent-url=your-agent-host:50051 -description="Developer Keys"
```

### 3. Deploy with Docker Compose (Recommended for Development)

```bash
# Setup and start all services
./scripts/setup-agent.sh setup

# Test the deployment
go run cmd/db-client/main.go -command=ping -agent=localhost:50051
```

### 4. Deploy to Kubernetes (Production)

```bash
# Set environment variables
export REGISTRY=your-registry.com
export IMAGE_TAG=v1.0.0
export DOMAIN=db-agent.yourdomain.com

# Deploy to Kubernetes
./scripts/deploy-k8s.sh deploy
```

## Deployment Options

### Option 1: Docker Compose (Development/Testing)

**Best for**: Development, testing, small teams

**Features**:
- Easy setup with single command
- Includes monitoring stack (Prometheus + Grafana)
- Local development environment
- Automatic health checks

**Setup**:
```bash
cd backend
./scripts/setup-agent.sh setup
```

**Access Points**:
- Agent gRPC: `localhost:50051`
- Metadata DB: `localhost:5433`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)

### Option 2: Kubernetes (Production)

**Best for**: Production, scalability, enterprise

**Features**:
- High availability with multiple replicas
- Auto-scaling based on load
- Rolling updates
- Service mesh integration
- Advanced monitoring and logging

**Setup**:
```bash
# Configure environment
export REGISTRY=your-registry.com
export DOMAIN=db-agent.yourdomain.com
export TLS_ENABLED=true

# Deploy
./scripts/deploy-k8s.sh deploy
```

### Option 3: Cloud Run (Google Cloud)

**Best for**: Serverless, pay-per-use, simple scaling

**Features**:
- Automatic scaling to zero
- Pay only for usage
- Built-in load balancing
- Automatic HTTPS

**Setup**:
```bash
# Build and push image
docker build -f Dockerfile.agent -t gcr.io/PROJECT-ID/db-agent:latest .
docker push gcr.io/PROJECT-ID/db-agent:latest

# Deploy to Cloud Run
gcloud run deploy db-agent \
  --image gcr.io/PROJECT-ID/db-agent:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 50051 \
  --set-env-vars METADB_HOST=your-cloud-sql-ip
```

### Option 4: AWS ECS/Fargate

**Best for**: AWS ecosystem, container orchestration

**Features**:
- Managed containers
- Integration with AWS services
- Auto-scaling
- Load balancing

**Setup**:
```bash
# Build and push to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT.dkr.ecr.us-east-1.amazonaws.com
docker build -f Dockerfile.agent -t db-agent:latest .
docker tag db-agent:latest ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/db-agent:latest
docker push ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/db-agent:latest

# Deploy using ECS CLI or AWS Console
```

### Option 5: Azure Container Instances

**Best for**: Azure ecosystem, simple container deployment

**Features**:
- Serverless containers
- Quick deployment
- Integration with Azure services

**Setup**:
```bash
# Build and push to ACR
az acr build --registry myregistry --image db-agent:latest .

# Deploy to ACI
az container create \
  --resource-group myResourceGroup \
  --name db-agent \
  --image myregistry.azurecr.io/db-agent:latest \
  --ports 50051 \
  --environment-variables METADB_HOST=your-sql-server.database.windows.net
```

## Security Configuration

### 1. TLS/SSL Setup

#### Self-Signed Certificate (Development)
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key \
  -out server.crt \
  -subj "/CN=your-agent-host"
```

#### Let's Encrypt (Production)
```bash
# Using certbot
certbot certonly --standalone -d your-agent-host.com

# Certificates will be in /etc/letsencrypt/live/your-agent-host.com/
```

### 2. API Key Authentication

```bash
# Generate secure API keys
go run cmd/db-keygen/main.go -server -description="Production Keys"

# Keys will be generated in keys/server/ directory
# Distribute keys/client/ directory to authorized users
```

### 3. Network Security

#### Firewall Rules
```bash
# AWS Security Group
aws ec2 authorize-security-group-ingress \
  --group-id sg-12345678 \
  --protocol tcp \
  --port 50051 \
  --cidr 10.0.0.0/8

# GCP Firewall
gcloud compute firewall-rules create allow-db-agent \
  --allow tcp:50051 \
  --source-ranges 10.0.0.0/8 \
  --target-tags db-agent
```

#### VPC Configuration
- Use private subnets for agent instances
- Restrict database access to agent only
- Enable VPC flow logs for monitoring

## Database Setup

### 1. Metadata Database

The agent requires a PostgreSQL database for tracking deployment history and metadata.

#### Local Setup (Docker)
```bash
# Included in docker-compose.agent.yml
docker-compose -f docker-compose.agent.yml up -d metadata-db
```

#### Cloud Setup (AWS RDS)
```bash
aws rds create-db-instance \
  --db-instance-identifier db-agent-metadata \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --master-username db_agent_user \
  --master-user-password secure_password \
  --allocated-storage 20 \
  --vpc-security-group-ids sg-12345678
```

#### Cloud Setup (Google Cloud SQL)
```bash
gcloud sql instances create db-agent-metadata \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --root-password=secure_password
```

### 2. Target Database Registration

```bash
# Register a database for management
go run cmd/db-client/main.go -command=bootstrap \
  -manifest=./manifest \
  -database=production \
  -agent=your-agent-host:50051
```

## Monitoring and Observability

### 1. Health Checks

```bash
# Using built-in client
go run cmd/db-client/main.go -command=ping -agent=your-agent-host:50051

# Using grpc_health_probe
grpc_health_probe -addr=your-agent-host:50051
```

### 2. Metrics Collection

The agent exposes Prometheus metrics on port 9090:

```bash
# View metrics
curl http://your-agent-host:9090/metrics
```

Key metrics:
- `db_agent_requests_total` - Total requests
- `db_agent_request_duration_seconds` - Request duration
- `db_agent_deployments_total` - Total deployments
- `db_agent_migrations_total` - Total migrations

### 3. Logging

#### Structured Logging
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "db-agent",
  "request_id": "req_123",
  "client_id": "client_456",
  "action": "deploy",
  "database": "production",
  "duration_ms": 1500
}
```

#### Log Aggregation
- **ELK Stack**: Elasticsearch, Logstash, Kibana
- **Fluentd**: Log forwarding and processing
- **Cloud Logging**: AWS CloudWatch, Google Cloud Logging, Azure Monitor

### 4. Alerting

#### Prometheus Alerts
```yaml
groups:
- name: db-agent
  rules:
  - alert: DatabaseAgentDown
    expr: up{job="db-agent"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Database Agent is down"
  
  - alert: HighErrorRate
    expr: rate(db_agent_requests_total{status="error"}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
```

## Backup and Disaster Recovery

### 1. Metadata Database Backup

```bash
# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h $METADB_HOST -U $METADB_USER $METADB_NAME > backup_$DATE.sql

# Upload to S3
aws s3 cp backup_$DATE.sql s3://your-backup-bucket/db-agent/
```

### 2. Configuration Backup

```bash
# Backup keys and configuration
tar -czf agent-config-backup.tar.gz keys/ config/
```

### 3. Disaster Recovery Plan

1. **RTO (Recovery Time Objective)**: 15 minutes
2. **RPO (Recovery Point Objective)**: 1 hour

**Recovery Steps**:
1. Restore metadata database from backup
2. Redeploy agent from latest image
3. Restore API keys and configuration
4. Verify connectivity and functionality

## Performance Tuning

### 1. Resource Allocation

#### Kubernetes
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

#### Docker
```bash
docker run -d \
  --name db-agent \
  --memory=1g \
  --cpus=1 \
  -p 50051:50051 \
  db-agent:latest
```

### 2. Connection Pooling

```go
// Agent configuration
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 3. Scaling

#### Horizontal Scaling
- Deploy multiple agent instances
- Use load balancer for distribution
- Implement session affinity if needed

#### Vertical Scaling
- Increase CPU and memory resources
- Optimize database connections
- Enable connection pooling

## Troubleshooting

### Common Issues

#### 1. Connection Refused
```bash
# Check if agent is running
netstat -tlnp | grep 50051

# Check firewall rules
iptables -L | grep 50051

# Check service status
systemctl status db-agent
```

#### 2. TLS Certificate Issues
```bash
# Verify certificate
openssl x509 -in server.crt -text -noout

# Test TLS connection
openssl s_client -connect your-agent-host:50051
```

#### 3. Database Connection Issues
```bash
# Test metadata database connection
psql -h $METADB_HOST -U $METADB_USER -d $METADB_NAME -c "SELECT 1;"

# Check database logs
tail -f /var/log/postgresql/postgresql.log
```

#### 4. Authentication Failures
```bash
# Verify API key
go run cmd/db-keygen/main.go -verify -key=your-api-key

# Check key permissions
ls -la keys/server/
```

### Debug Mode

```bash
# Run agent with debug logging
go run cmd/db-agent/main.go -port=50051 -log-level=debug

# Test client connection with verbose output
go run cmd/db-client/main.go -command=ping -agent=localhost:50051 -verbose
```

## Maintenance

### 1. Updates

```bash
# Update agent
docker pull your-registry/db-agent:latest
docker-compose -f docker-compose.agent.yml up -d db-agent

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
# Monitor for authentication failures
```

### 3. Database Maintenance

```bash
# Vacuum and analyze metadata database
psql -h $METADB_HOST -U $METADB_USER -d $METADB_NAME -c "VACUUM ANALYZE;"

# Check database size
psql -h $METADB_HOST -U $METADB_USER -d $METADB_NAME -c "SELECT pg_size_pretty(pg_database_size('db_agent_metadata'));"
```

## Cost Optimization

### 1. Resource Optimization

- Use appropriate instance sizes
- Enable auto-scaling
- Implement connection pooling
- Optimize database queries

### 2. Storage Optimization

- Compress logs and backups
- Implement log rotation
- Use appropriate storage classes
- Clean up old data regularly

### 3. Network Optimization

- Use private networks when possible
- Implement caching
- Optimize API calls
- Use CDN for static assets

## Security Best Practices

### 1. Access Control

- Use least privilege principle
- Implement API key rotation
- Enable audit logging
- Monitor access patterns

### 2. Network Security

- Use TLS for all connections
- Implement firewall rules
- Use private networks
- Enable VPC flow logs

### 3. Data Protection

- Encrypt data at rest
- Encrypt data in transit
- Implement backup encryption
- Use secure key management

## Support and Documentation

### 1. Getting Help

- Check logs: `docker-compose -f docker-compose.agent.yml logs db-agent`
- Test connectivity: `go run cmd/db-client/main.go -command=ping -agent=your-host:50051`
- Verify configuration: Check environment variables and config files

### 2. Documentation

- [Agent README](cmd/db-agent/README.md)
- [Client Usage Guide](cmd/db-client/README.md)
- [API Reference](proto/dbagent/README.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)

### 3. Community

- GitHub Issues: Report bugs and request features
- Discussions: Ask questions and share experiences
- Contributing: Submit pull requests and improvements

This guide provides comprehensive information for deploying the Database Agent to various cloud environments. Choose the deployment method that best fits your requirements and infrastructure.
