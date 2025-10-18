# Terraform State Backup

## Critical State Information

The following information has been extracted from the Terraform state and should be backed up to GCloud Secret Manager:

### Terraform Outputs (terraform-outputs.json)
```json
{
  "artifact_registry_repo": "civicweave",
  "backend_url": "https://civicweave-backend-peedoie7va-uc.a.run.app",
  "database_connection_name": "civicweave-474622:us-central1:civicweave-postgres",
  "frontend_url": "https://civicweave-frontend-peedoie7va-uc.a.run.app",
  "project_id": "civicweave-474622",
  "redis_host": "10.63.105.59",
  "redis_port": 6379
}
```

### Backup to GCloud Secret Manager

Run the following commands to backup the state:

```bash
# Backup Terraform outputs
gcloud secrets create terraform-outputs-backup --data-file=terraform-outputs.json --project=civicweave-474622

# Backup full state (if needed for complex recovery)
gcloud secrets create terraform-state-full-backup --data-file=terraform.tfstate --project=civicweave-474622
```

### Recovery Process

To restore state information for redeployment:

```bash
# Retrieve outputs
gcloud secrets versions access latest --secret="terraform-outputs-backup" --project=civicweave-474622 > terraform-outputs.json

# Use outputs in deployment scripts
export BACKEND_URL=$(jq -r '.backend_url.value' terraform-outputs.json)
export FRONTEND_URL=$(jq -r '.frontend_url.value' terraform-outputs.json)
export DATABASE_CONNECTION_NAME=$(jq -r '.database_connection_name.value' terraform-outputs.json)
export REDIS_HOST=$(jq -r '.redis_host.value' terraform-outputs.json)
```

### Migration to Remote Backend

When ready to migrate to a remote backend:

1. Create a GCS bucket for state storage
2. Update `main.tf` with remote backend configuration
3. Run `terraform init -migrate-state` to migrate state to remote backend
4. Remove local state files after successful migration

## Files to Remove from Git

- `terraform.tfstate`
- `terraform.tfstate.backup`
- `terraform-outputs.json` (temporary backup file)
