package analyse

import (
        "github.com/chaosinthecrd/dexter/pkg/files"
        "github.com/chaosinthecrd/dexter/pkg/config"
)

type Analyser struct {
   Ignores []config.Ignore
   Parsers []files.Parser
}

// NewAnalyzer generate a NewAnalyzer with rules to apply.
func NewAnalyser(configFile string) (Analyser, error) {
   analyser := Analyser{}
   conf, err := config.InitialiseConfig(configFile)
   if err != nil {
      return Analyser{}, err
   }

   analyser.Ignores = conf.Ignores


   return analyser, nil
}
