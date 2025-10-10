# GCP Project Configuration
project_id   = "civicweave-474622"
project_name = "civicweave"
region       = "us-central1"
zone         = "us-central1-a"

# Security (CHANGE THESE VALUES!)
jwt_secret = "changeme-jwt-secret"

# Email Service (set via environment variables: TF_VAR_mailgun_api_key, TF_VAR_mailgun_domain)
# mailgun_api_key = "set via TF_VAR_mailgun_api_key"
# mailgun_domain  = "set via TF_VAR_mailgun_domain"

# Google OAuth (set via environment variables: TF_VAR_google_client_id, TF_VAR_google_client_secret)
# google_client_id     = "set via TF_VAR_google_client_id"
# google_client_secret = "set via TF_VAR_google_client_secret"

# Database (CHANGE THIS PASSWORD!)
db_password = "changeme-db-password"
