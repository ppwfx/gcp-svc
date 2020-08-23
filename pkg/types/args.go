package types

type IntegrationTestArgs struct {
	UserSvcAddr string
	PostgresUrl string
	Remote      bool
}

type ServeArgs struct {
	PostgresUrl            string
	Port                   string
	HmacSecret             string
	AllowedSubjectSuffix   string
	Metrics                string
	Logging                string
	Migrate                string
	ExposePprof            bool
	HttpReadTimeoutSeconds int
}

const (
	MetricsStackDriver = "stackdriver"
	LoggingStackDriver = "stackdriver"
)
