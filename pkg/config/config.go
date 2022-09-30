package config

import (
        "gopkg.in/yaml.v3"
        "fmt"
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

   fmt.Printf("Finding Dexter Config")
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

      fmt.Printf("DexterConfig: %s \n", config)

   }

   return config, nil
}

