package config

import (
        "gopkg.in/yaml.v3"
	"os"
)

type DexterConfig struct {
   Ignores []Ignore `yaml:"ignore"`
   Parsers []string `yaml:"parsers"`
}

type Ignore struct {
   File string
   Reference string
}


func InitialiseConfig(configFile string) (DexterConfig, error) {

   config := DexterConfig{}
   if configFile != "" {
      file, err := os.ReadFile(configFile)
      if err != nil {
         return DexterConfig{}, err
      }

      err = yaml.Unmarshal(file, &config)
      if err != nil {
         return DexterConfig{}, err
      }

   }

   return config, nil
}

