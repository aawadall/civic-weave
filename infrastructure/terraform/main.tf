terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# Configure the Google Cloud Provider
provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "redis.googleapis.com",
    "secretmanager.googleapis.com",
    "cloudbuild.googleapis.com",
    "artifactregistry.googleapis.com"
  ])

  project = var.project_id
  service = each.value

  disable_dependent_services = false
  disable_on_destroy         = false
}

# Cloud SQL (PostgreSQL) instance
resource "google_sql_database_instance" "postgres" {
  name             = "${var.project_name}-postgres"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro" # Smallest instance for cost optimization

    backup_configuration {
      enabled    = true
      start_time = "03:00"
    }

    ip_configuration {
      ipv4_enabled    = true  # Keep public IP for now - can be made private later
      ssl_mode        = "ENCRYPTED_ONLY"
      authorized_networks {
        name  = "cloud-run"
        value = "0.0.0.0/0"  # Allow Cloud Run to connect
      }
    }
  }

  deletion_protection = false # Set to true in production

  depends_on = [google_project_service.required_apis]
}

# Create database
resource "google_sql_database" "civicweave_db" {
  name     = "civicweave"
  instance = google_sql_database_instance.postgres.name
}

# Create database user
resource "google_sql_user" "civicweave_user" {
  name     = "civicweave"
  instance = google_sql_database_instance.postgres.name
  password = var.db_password
}

# Memorystore (Redis) instance
resource "google_redis_instance" "redis" {
  name           = "${var.project_name}-redis"
  tier           = "BASIC"
  memory_size_gb = 1
  region         = var.region

  depends_on = [google_project_service.required_apis]
}

# Secret Manager secrets
resource "google_secret_manager_secret" "secrets" {
  for_each = toset([
    "jwt-secret",
    "mailgun-api-key",
    "mailgun-domain",
    "google-client-id",
    "google-client-secret",
    "db-password",
    "admin-password",
    "openai-api-key"
  ])

  secret_id = each.key

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }

  depends_on = [google_project_service.required_apis]
}

# Secret versions
resource "google_secret_manager_secret_version" "jwt_secret" {
  secret      = google_secret_manager_secret.secrets["jwt-secret"].id
  secret_data = var.jwt_secret
}

resource "google_secret_manager_secret_version" "mailgun_api_key" {
  secret      = google_secret_manager_secret.secrets["mailgun-api-key"].id
  secret_data = var.mailgun_api_key
}

resource "google_secret_manager_secret_version" "mailgun_domain" {
  secret      = google_secret_manager_secret.secrets["mailgun-domain"].id
  secret_data = var.mailgun_domain
}

resource "google_secret_manager_secret_version" "google_client_id" {
  secret      = google_secret_manager_secret.secrets["google-client-id"].id
  secret_data = var.google_client_id
}

resource "google_secret_manager_secret_version" "google_client_secret" {
  secret      = google_secret_manager_secret.secrets["google-client-secret"].id
  secret_data = var.google_client_secret
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.secrets["db-password"].id
  secret_data = var.db_password
}

resource "google_secret_manager_secret_version" "admin_password" {
  secret      = google_secret_manager_secret.secrets["admin-password"].id
  secret_data = var.admin_password
}

resource "google_secret_manager_secret_version" "openai_api_key" {
  secret      = google_secret_manager_secret.secrets["openai-api-key"].id
  secret_data = var.openai_api_key
}

# Service account for Cloud Run backend
resource "google_service_account" "civicweave_sa" {
  account_id   = "${var.project_name}-backend-sa"
  display_name = "CivicWeave Backend Service Account"
}

# Service account for Cloud Run frontend
resource "google_service_account" "civicweave_frontend_sa" {
  account_id   = "${var.project_name}-frontend-sa"
  display_name = "CivicWeave Frontend Service Account"
}

# IAM bindings for backend service account
resource "google_project_iam_member" "backend_sa_permissions" {
  for_each = toset([
    "roles/secretmanager.secretAccessor",
    "roles/cloudsql.client",
    "roles/redis.viewer"
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.civicweave_sa.email}"
}

# IAM bindings for frontend service account (minimal permissions)
resource "google_project_iam_member" "frontend_sa_permissions" {
  for_each = toset([
    # Frontend only needs basic Cloud Run permissions
    # No access to secrets, database, or other sensitive resources
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.civicweave_frontend_sa.email}"
}

# Grant Cloud Run invoker permission to the service account
resource "google_cloud_run_service_iam_member" "backend_invoker" {
  location = google_cloud_run_v2_service.backend.location
  service  = google_cloud_run_v2_service.backend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_service_iam_member" "frontend_invoker" {
  location = google_cloud_run_v2_service.frontend.location
  service  = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Cloud Run service for backend
resource "google_cloud_run_v2_service" "backend" {
  name     = "${var.project_name}-backend"
  location = var.region

  template {
    service_account = google_service_account.civicweave_sa.email

    # Cloud SQL connection via Unix socket
    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.postgres.connection_name]
      }
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.project_name}/backend:latest"

      env {
        name  = "DB_HOST"
        value = "/cloudsql/${google_sql_database_instance.postgres.connection_name}"
      }
      env {
        name  = "DB_PORT"
        value = "5432"
      }
      env {
        name  = "DB_NAME"
        value = google_sql_database.civicweave_db.name
      }
      env {
        name  = "DB_USER"
        value = google_sql_user.civicweave_user.name
      }
      env {
        name = "DB_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["db-password"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name  = "REDIS_HOST"
        value = google_redis_instance.redis.host
      }
      env {
        name  = "REDIS_PORT"
        value = tostring(google_redis_instance.redis.port)
      }
      env {
        name = "JWT_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["jwt-secret"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "MAILGUN_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["mailgun-api-key"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "MAILGUN_DOMAIN"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["mailgun-domain"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "GOOGLE_CLIENT_ID"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["google-client-id"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "GOOGLE_CLIENT_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["google-client-secret"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name  = "NOMINATIM_BASE_URL"
        value = "https://nominatim.openstreetmap.org"
      }
      env {
        name  = "ADMIN_EMAIL"
        value = "admin@civicweave.com"
      }
      env {
        name  = "ADMIN_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["admin-password"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name  = "ADMIN_NAME"
        value = "System Administrator"
      }
      env {
        name = "OPENAI_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.secrets["openai-api-key"].secret_id
            version = "latest"
          }
        }
      }
      env {
        name  = "OPENAI_EMBEDDING_MODEL"
        value = var.openai_embedding_model
      }
      env {
        name  = "ENABLE_EMAIL"
        value = "false"
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }

  depends_on = [google_project_service.required_apis]
}

# Cloud Run service for frontend
resource "google_cloud_run_v2_service" "frontend" {
  name     = "${var.project_name}-frontend"
  location = var.region

  template {
    service_account = google_service_account.civicweave_frontend_sa.email

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.project_name}/frontend:latest"

      env {
        name  = "VITE_API_BASE_URL"
        value = "${google_cloud_run_v2_service.backend.uri}/api"
      }
      env {
        name  = "VITE_GOOGLE_CLIENT_ID"
        value = var.google_client_id
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }

  depends_on = [google_project_service.required_apis]
}

# Artifact Registry for container images
resource "google_artifact_registry_repository" "civicweave_repo" {
  location      = var.region
  repository_id = var.project_name
  description   = "CivicWeave container repository"
  format        = "DOCKER"

  depends_on = [google_project_service.required_apis]
}
