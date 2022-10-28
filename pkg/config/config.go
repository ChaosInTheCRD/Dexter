package config

import (
        "gopkg.in/yaml.v3"
	"os"
)

type DexterConfig struct {
   Ignores Ignores `yaml:"ignore"`
   Parsers []string `yaml:"parsers"`
}

type Ignores struct {
   Files []string `yaml:"files"`
   References []string `yaml:"references"`
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

