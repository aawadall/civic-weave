# Deploy Database Agent to Google Cloud Run

## Quick Start

### 1. Prerequisites

- Google Cloud account with billing enabled
- `gcloud` CLI installed and authenticated
- Go 1.21+ installed locally

### 2. Setup

```bash
# Install gcloud CLI (if not already installed)
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# Authenticate with Google Cloud
gcloud auth login
gcloud auth application-default login

# Set your project ID
export PROJECT_ID=your-project-id
```

### 3. Deploy

```bash
cd backend

# Deploy everything with one command
./scripts/deploy-gcloud.sh deploy
```

That's it! The script will:
- Enable required APIs
- Create a Cloud SQL PostgreSQL instance for metadata
- Build and push the Docker image
- Deploy to Cloud Run
- Set up the database schema
- Create API keys for authentication

### 4. Test

```bash
# Get your service URL (shown after deployment)
export DB_AGENT_URL=https://db-agent-xxx-uc.a.run.app:443
export DB_AGENT_API_KEY=your_api_key_from_deployment

# Test the connection
go run cmd/db-client/main.go -command=ping -agent=$DB_AGENT_URL
```

## What Gets Deployed

### Cloud Run Service
- **Name**: `db-agent`
- **Port**: 50051 (gRPC)
- **Memory**: 1GB
- **CPU**: 1 vCPU
- **Scaling**: 0-10 instances (auto-scaling)
- **Timeout**: 5 minutes

### Cloud SQL Database
- **Instance**: `db-agent-metadata`
- **Database**: `db_agent_metadata`
- **User**: `db_agent_user`
- **Tier**: `db-f1-micro` (can be upgraded)
- **Region**: Same as Cloud Run service

### Security
- **Authentication**: API key-based
- **Network**: Private Cloud SQL connection
- **Secrets**: Stored in Google Secret Manager

## Commands

### Deploy
```bash
./scripts/deploy-gcloud.sh deploy
```

### Update
```bash
./scripts/deploy-gcloud.sh update
```

### Check Status
```bash
./scripts/deploy-gcloud.sh status
```

### View Logs
```bash
./scripts/deploy-gcloud.sh logs
```

### Cleanup
```bash
./scripts/deploy-gcloud.sh cleanup
```

## Configuration

### Environment Variables

You can customize the deployment:

```bash
export PROJECT_ID=your-project-id
export REGION=us-central1          # or europe-west1, asia-southeast1
export SERVICE_NAME=db-agent       # Cloud Run service name
export IMAGE_TAG=v1.0.0           # Docker image tag
```

### Makefile Integration

Add to your Makefile:

```makefile
# Google Cloud Run Deployment
db-agent-gcloud:
	cd backend && ./scripts/deploy-gcloud.sh deploy

db-agent-gcloud-update:
	cd backend && ./scripts/deploy-gcloud.sh update

db-agent-gcloud-logs:
	cd backend && ./scripts/deploy-gcloud.sh logs

db-agent-gcloud-cleanup:
	cd backend && ./scripts/deploy-gcloud.sh cleanup
```

## Usage Examples

### 1. Test Connection
```bash
go run cmd/db-client/main.go -command=ping -agent=$DB_AGENT_URL
```

### 2. Compare Manifest
```bash
go run cmd/db-client/main.go -command=compare \
  -manifest=./manifest \
  -database=production \
  -agent=$DB_AGENT_URL
```

### 3. Deploy Manifest (Dry Run)
```bash
go run cmd/db-client/main.go -command=deploy \
  -manifest=./manifest \
  -database=production \
  -agent=$DB_AGENT_URL \
  -dry-run
```

### 4. Deploy Manifest (Actual)
```bash
go run cmd/db-client/main.go -command=deploy \
  -manifest=./manifest \
  -database=production \
  -agent=$DB_AGENT_URL
```

### 5. Download Current Schema
```bash
go run cmd/db-client/main.go -command=download \
  -database=production \
  -output=./current-schema \
  -agent=$DB_AGENT_URL
```

## Cost Estimation

### Cloud Run
- **Free tier**: 2 million requests/month
- **Pricing**: $0.40 per million requests + compute time
- **Estimated cost**: $5-20/month for moderate usage

### Cloud SQL
- **db-f1-micro**: ~$7/month
- **db-g1-small**: ~$25/month (recommended for production)
- **Storage**: $0.17/GB/month

### Total Estimated Cost
- **Development**: ~$10/month
- **Production**: ~$30-50/month

## Monitoring

### Cloud Run Metrics
- Request count
- Request latency
- Error rate
- Instance count

### Cloud SQL Metrics
- CPU utilization
- Memory usage
- Connection count
- Storage usage

### Application Logs
```bash
# View logs in real-time
gcloud logging tail "resource.type=cloud_run_revision AND resource.labels.service_name=db-agent"

# View logs in Cloud Console
# https://console.cloud.google.com/run
```

## Troubleshooting

### Common Issues

#### 1. Authentication Error
```bash
# Check if API key is correct
echo $DB_AGENT_API_KEY

# Test with verbose output
go run cmd/db-client/main.go -command=ping -agent=$DB_AGENT_URL -verbose
```

#### 2. Connection Timeout
```bash
# Check Cloud Run service status
gcloud run services describe db-agent --region=us-central1

# Check logs
./scripts/deploy-gcloud.sh logs
```

#### 3. Database Connection Issues
```bash
# Check Cloud SQL instance
gcloud sql instances describe db-agent-metadata

# Test database connection
gcloud sql connect db-agent-metadata --user=db_agent_user --database=db_agent_metadata
```

### Debug Mode

```bash
# Deploy with debug logging
gcloud run services update db-agent \
  --region=us-central1 \
  --set-env-vars LOG_LEVEL=debug
```

## Security Best Practices

### 1. API Key Management
- Store API keys in environment variables
- Rotate keys regularly
- Use different keys for different environments

### 2. Network Security
- Cloud SQL is only accessible from Cloud Run
- No public IP addresses
- All traffic encrypted in transit

### 3. Access Control
- Use IAM roles for service accounts
- Principle of least privilege
- Enable audit logging

## Scaling

### Auto-scaling
Cloud Run automatically scales based on:
- Request volume
- CPU utilization
- Memory usage

### Manual Scaling
```bash
# Set minimum instances
gcloud run services update db-agent \
  --region=us-central1 \
  --min-instances=1

# Set maximum instances
gcloud run services update db-agent \
  --region=us-central1 \
  --max-instances=50
```

## Backup and Recovery

### Database Backups
Cloud SQL automatically creates backups:
- Daily backups (retained for 7 days)
- Weekly backups (retained for 4 weeks)
- Monthly backups (retained for 12 months)

### Manual Backup
```bash
# Create manual backup
gcloud sql backups create \
  --instance=db-agent-metadata \
  --description="Manual backup before deployment"
```

### Recovery
```bash
# Restore from backup
gcloud sql backups restore BACKUP_ID \
  --restore-instance=db-agent-metadata
```

## Production Considerations

### 1. Upgrade Database Tier
```bash
gcloud sql instances patch db-agent-metadata \
  --tier=db-g1-small
```

### 2. Enable SSL
```bash
gcloud sql instances patch db-agent-metadata \
  --require-ssl
```

### 3. Set up Monitoring
- Enable Cloud Monitoring
- Set up alerting policies
- Configure log-based metrics

### 4. Custom Domain
```bash
# Map custom domain
gcloud run domain-mappings create \
  --service=db-agent \
  --domain=db-agent.yourdomain.com \
  --region=us-central1
```

This guide provides everything you need to deploy the Database Agent to Google Cloud Run quickly and easily!
