terraform {
  backend "gcs" {
    bucket  = "tf-user-svc"
    prefix  = "terraform/state"
    credentials = "credentials.json"
  }
}

provider "google" {
  project     = "user-svc"
  credentials = file("credentials.json")
  region      = "us-east1"
  zone        = "us-east1-a"
}

provider "google-beta" {
  project     = "user-svc"
  credentials = file("credentials.json")
  region      = "us-east1"
  zone        = "us-east1-a"
}

resource "google_sql_database_instance" "postgresql" {
  name = "user-svc-0"
  database_version = "POSTGRES_12"

  settings {
    tier = "db-f1-micro"
    activation_policy = "ALWAYS"
    disk_autoresize = true
    disk_size = 10
    disk_type = "PD_SSD"
    pricing_plan = "PER_USE"

    # sunday 3am
    maintenance_window {
      day  = "7"
      hour = "3"
    }

    backup_configuration {
      enabled = true
      start_time = "00:00"
    }

    ip_configuration {
      ipv4_enabled = "true"
      authorized_networks {
        value = "0.0.0.0/0"
      }
    }
  }
}

resource "google_sql_database" "postgresql_db" {
  instance = google_sql_database_instance.postgresql.name
  name = "user-svc"
}

resource "random_id" "user_password" {
  byte_length = 32
}
resource "google_sql_user" "postgresql_user" {
  instance = google_sql_database_instance.postgresql.name
  name = "root"
  password = random_id.user_password.hex
}

resource "google_cloud_run_service" "user-svc" {
  name     = "user-svc"
  location = "us-east1"
  autogenerate_revision_name = true

  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale"      = "10"
        "run.googleapis.com/client-name"        = "terraform"
        "run.googleapis.com/cloudsql-instances" = "user-svc:us-east1:${google_sql_database_instance.postgresql.name}"
      }
    }

    spec {
      containers {
        image = "gcr.io/user-svc/user-svc:latest"
        command = ["./user-svc"]
        args = ["--addr", "0.0.0.0:8080", "--db-connection", "host=/cloudsql/${google_sql_database_instance.postgresql.connection_name} user=${google_sql_user.postgresql_user.name} password=${google_sql_user.postgresql_user.password} dbname=${google_sql_database.postgresql_db.name} sslmode=disable", "--hmac-secret", "x", "--allowed-subject-suffix", "@test.com"]
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

data "google_iam_policy" "public" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "user-svc-public" {
  location    = google_cloud_run_service.user-svc.location
  service     = google_cloud_run_service.user-svc.name

  policy_data = data.google_iam_policy.public.policy_data
}

