package server

type Config struct {
	CacheUrl   string `toml:"cacheUrl"`
	StoreUrl   string `toml:"storeUrl"`
	ServerAddr string `toml:"serverAddr"`
	BasePath   string `toml:"basePath"`
	BaseUrl    string `toml:"baseUrl"`
}

func NewConfig() *Config {
	return &Config{
		CacheUrl:   "redis://user:@localhost:6379/1",
		StoreUrl:   "user=postgres password=Tylpa31 dbname=cataloguer_test sslmode=disable",
		ServerAddr: "127.0.0.1:8080",
		BasePath:   "C:\\Users\\User\\GolandProjects\\Cataloguer",
		BaseUrl:    "/account/messages/",
	}
}
