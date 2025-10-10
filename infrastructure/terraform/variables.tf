variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "project_name" {
  description = "The name prefix for resources"
  type        = string
  default     = "civicweave"
}

variable "region" {
  description = "The GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The GCP zone"
  type        = string
  default     = "us-central1-a"
}

variable "jwt_secret" {
  description = "JWT secret key"
  type        = string
  sensitive   = true
}

variable "mailgun_api_key" {
  description = "Mailgun API key"
  type        = string
  sensitive   = true
}

variable "mailgun_domain" {
  description = "Mailgun domain"
  type        = string
}

variable "google_client_id" {
  description = "Google OAuth client ID"
  type        = string
}

variable "google_client_secret" {
  description = "Google OAuth client secret"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "admin_password" {
  description = "Admin user password"
  type        = string
  sensitive   = true
}

variable "openai_api_key" {
  description = "OpenAI API key for skill embeddings"
  type        = string
  sensitive   = true
}

variable "openai_embedding_model" {
  description = "OpenAI embedding model"
  type        = string
  default     = "text-embedding-3-small"
}
