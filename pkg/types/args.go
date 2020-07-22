package types

type IntegrationTestArgs struct {
	UserSvcAddr  string
	DbConnection string
	Remote       bool
}

type ServeArgs struct {
	DbConnection         string
	Addr                 string
	HmacSecret           string
	AllowedSubjectSuffix string
}
