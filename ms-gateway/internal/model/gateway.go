package model

type Gateway struct {
	Routes []Route `yaml:"routes,omitempty"`
}

type Route struct {
	Id         string   `yaml:"id,omitempty"`
	Uri        string   `yaml:"uri,omitempty"`
	PathPrefix string   `yaml:"pathPrefix,omitempty"`
	Filters    []string `yaml:"filters,omitempty"`
}
