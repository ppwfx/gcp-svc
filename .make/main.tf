variable "user-svc-version" {
  type = string
  default = "skip"
}

terraform {
  backend "gcs" {
    bucket = "tf-user-svc"
    prefix = "terraform/state"
    credentials = "credentials.json"
  }
}

provider "google" {
  project = "user-svc"
  credentials = file("credentials.json")
  region = "us-east1"
  zone = "us-east1-a"
}

provider "google-beta" {
  project = "user-svc"
  credentials = file("credentials.json")
  region = "us-east1"
  zone = "us-east1-a"
}

module "user-svc-postgres" {
  source = "./modules/user-svc-postgres"
}

module "user-svc" {
  source = "./modules/user-svc"

  user-svc-version = var.user-svc-version
  postgresql_instance_connection_name = module.user-svc-postgres.connection_name
  container_args = [
    "--addr", "0.0.0.0:8080",
    "--db-connection", module.user-svc-postgres.sqlx_connection_string,
    "--hmac-secret", "x",
    "--allowed-subject-suffix", "@test.com",
    "--metrics", "stackdriver",
    "--logging", "stackdriver",
  ]
}

output "user-svc-url" {
  value = module.user-svc.url
}