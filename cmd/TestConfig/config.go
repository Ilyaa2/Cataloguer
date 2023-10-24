package TestConfig

type Config struct {
	CacheUrl   string
	StoreUrl   string
	ServerAddr string
	BasePath   string
}

func NewConfig() *Config {
	return &Config{
		CacheUrl:   "redis://user:@localhost:6379/1",
		StoreUrl:   "user=postgres password=Tylpa31 dbname=cataloguer_test sslmode=disable",
		ServerAddr: "127.0.0.1:8080",
		BasePath:   "C:\\Users\\User\\GolandProjects\\Cataloguer",
	}
}
