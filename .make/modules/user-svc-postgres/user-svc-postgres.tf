resource "google_sql_database_instance" "postgresql" {
  name = "user-svc-2"
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

output "connection_name" {
  value = google_sql_database_instance.postgresql.connection_name
}

output "sqlx_connection_string" {
  value = "host=/cloudsql/${google_sql_database_instance.postgresql.connection_name} user=${google_sql_user.postgresql_user.name} password=${google_sql_user.postgresql_user.password} dbname=${google_sql_database.postgresql_db.name} sslmode=disable"
}

output "connection_url" {
  value = "postgres://${google_sql_user.postgresql_user.name}:${google_sql_user.postgresql_user.password}@${google_sql_database_instance.postgresql.ip_address.0.ip_address}:5432/${google_sql_database.postgresql_db.name}"
}