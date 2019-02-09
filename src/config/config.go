package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	nextcloud "../nextcloud"
	"errors"
)

type ConfigPaths struct {
	Videos string `yaml:"videos"`
	Audios string `yaml:"audios"`
}

type ConfigOutput struct {
	Paths ConfigPaths `yaml:"paths"`
}

type Config struct {
	Output ConfigOutput `yaml:"output"`
}

func GetDefaultConfig(path string) (Config, error) {
	cfg := Config{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func FromBytes(bytes []byte) (Config, error) {
	cfg := Config{}

	err := yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func IsValid(cfg Config, nextCloudClient *nextcloud.NextCloudClient) error {
	if cfg.Output.Paths.Videos == "" {
		return errors.New("output/paths/videos empty in config.yml")
	}

	if cfg.Output.Paths.Audios == "" {
		return errors.New("output/paths/audios empty in config.yml")
	}

	exists, err := nextCloudClient.FileExists(cfg.Output.Paths.Videos)
	if err != nil {
		return errors.New("Couldn't check whether videos path exists: " + err.Error())
	}

	if !exists {
		return errors.New("Path " + cfg.Output.Paths.Videos + " doesn't exist")
	}

	exists, err = nextCloudClient.FileExists(cfg.Output.Paths.Audios)
	if err != nil {
		return errors.New("Couldn't check whether audios path exists: " + err.Error())
	}

	if !exists {
		return errors.New("Path " + cfg.Output.Paths.Audios + " doesn't exist")
	}

	return nil
}