output "project_id" {
  description = "The GCP project ID"
  value       = var.project_id
}

output "backend_url" {
  description = "Backend Cloud Run service URL"
  value       = google_cloud_run_v2_service.backend.uri
}

output "frontend_url" {
  description = "Frontend Cloud Run service URL"
  value       = google_cloud_run_v2_service.frontend.uri
}

output "database_connection_name" {
  description = "Cloud SQL instance connection name"
  value       = google_sql_database_instance.postgres.connection_name
}

output "redis_host" {
  description = "Redis instance host"
  value       = google_redis_instance.redis.host
}

output "redis_port" {
  description = "Redis instance port"
  value       = google_redis_instance.redis.port
}

output "artifact_registry_repo" {
  description = "Artifact Registry repository URL"
  value       = google_artifact_registry_repository.civicweave_repo.name
}
