package lib

type ServerConfiguration struct {
	HostName string `yaml:"hostname"`
	Port     string `yaml:"port"`
}

type Configurations struct {
	Server ServerConfiguration
	// these are env variables
	EXAMPLE_VAR  string `yaml:"EXAMPLE_VAR"`
	EXAMPLE_PATH string `yaml:"EXAMPLE_PATH"`
}
