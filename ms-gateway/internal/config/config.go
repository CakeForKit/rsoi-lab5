package config

import "github.com/CakeForKit/rsoi-lab2/ms-gateway/internal/model"

var AppConfig Config

type Config struct {
	Application ApplicationProperty `yaml:"application,omitempty"`
}

type ApplicationProperty struct {
	Gateway model.Gateway `yaml:"gateway,omitempty"`
}
