package types

type IntegrationTestArgs struct {
	DbConnection string
	UserSvcAddr  string
}

type ServeArgs struct {
	DbConnection string
	Addr         string
	HmacSecret   string
	Salt         string
}
