variable "svc-version" {
  type = string
}

variable "postgresql_instance_connection_name" {
  type = string
}

variable "container_args" {
  type = list(string)
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
        "run.googleapis.com/cloudsql-instances" = var.postgresql_instance_connection_name
      }
    }

    spec {
      containers {
        image = "gcr.io/user-svc/user-svc:${var.svc-version}"
        command = ["./user-svc"]
        args = var.container_args
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

output "url" {
  value = "${trimprefix(google_cloud_run_service.user-svc.status[0].url, "https://")}:443"
}