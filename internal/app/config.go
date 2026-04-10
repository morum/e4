package app

type Config struct {
	ListenAddr  string
	HostKeyPath string
	LogLevel    string
}

func DefaultConfig() Config {
	return Config{
		ListenAddr:  ":2222",
		HostKeyPath: ".e4_host_key",
		LogLevel:    "info",
	}
}
