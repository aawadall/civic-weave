# CivicWeave Infrastructure Deployment

This directory contains the Terraform configuration for deploying CivicWeave to Google Cloud Platform.

## Prerequisites

1. **GCP Project Setup**: Run `make setup-gcp` from the project root
2. **Billing Enabled**: Ensure billing is enabled for your GCP project
3. **gcloud CLI**: Authenticated and configured
4. **Terraform**: Installed (>= 1.0)

## Environment Variables Setup

### 1. Copy the example environment file

```bash
cp env.example .env
```

### 2. Update .env with your actual values

Edit `.env` and replace the placeholder values:

```bash
# Mailgun Configuration (Required)
TF_VAR_mailgun_api_key=your-actual-mailgun-api-key
TF_VAR_mailgun_domain=your-actual-mailgun-domain.com

# Google OAuth Configuration (Optional for MVP)
TF_VAR_google_client_id=your-actual-google-oauth-client-id
TF_VAR_google_client_secret=your-actual-google-oauth-client-secret
```

### 3. Security Note

- **DO NOT** commit the `.env` file to version control
- The `.env` file is already in `.gitignore`
- Use strong, unique passwords and API keys

## Deployment

### Deploy Infrastructure

```bash
# From project root
make deploy-infra
```

This will:
- Load environment variables from `.env`
- Initialize Terraform
- Create all GCP resources (Cloud SQL, Redis, Cloud Run, etc.)

### Build and Push Images

```bash
# Set project ID (replace with your actual project ID)
export PROJECT_ID=civicweave-474622

# Build and push container images
make build-push
```

### Deploy Applications

```bash
make deploy-app
```

## What Gets Created

- **Cloud SQL**: PostgreSQL database (db-f1-micro for cost optimization)
- **Memorystore**: Redis cache (1GB basic tier)
- **Cloud Run**: Backend and frontend services
- **Secret Manager**: Secure storage for API keys and passwords
- **Artifact Registry**: Docker repository for container images
- **IAM**: Service accounts with appropriate permissions

## Cost Optimization

The configuration is optimized for cost:
- Smallest Cloud SQL instance (db-f1-micro)
- Basic Redis tier (1GB)
- Cloud Run with auto-scaling (0-10 instances)
- Regional deployment in us-central1

## Troubleshooting

### Permission Errors
If you get permission errors:
```bash
# Update Application Default Credentials
gcloud auth application-default set-quota-project civicweave-474622
```

### API Not Enabled
If an API fails to enable:
```bash
# Enable manually
gcloud services enable [API_NAME] --project=civicweave-474622
```

### Terraform State Issues
If Terraform state gets corrupted:
```bash
# Reinitialize (be careful - this can cause data loss)
terraform init -reconfigure
```

## Security Considerations

- All secrets are stored in Google Secret Manager
- Database uses SSL encryption
- Cloud Run services use service accounts with minimal permissions
- Public access is granted only to necessary services

## Monitoring and Logs

- **Cloud Run Logs**: Available in GCP Console > Cloud Run > [Service] > Logs
- **Cloud SQL Logs**: Available in GCP Console > SQL > [Instance] > Logs
- **Redis Monitoring**: Available in GCP Console > Memorystore > [Instance] > Monitoring