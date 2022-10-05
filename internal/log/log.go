package log

import (
	"context"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

type logKey struct{}

func InitLogContext(debug bool) (context.Context){
   var level log.Level

   if debug == true {
      level = log.DebugLevel
   } else {
      level = log.InfoLevel
   }

   logger := log.Logger{
        Handler: cli.New(os.Stdout),
        Level: level,
   }
    

    ctx := context.TODO() 
    ctx = log.NewContext(ctx, log.NewEntry(&logger))

    return ctx
}


func AddFields(logger log.Interface, command, directory, config string) (*log.Entry){

    entry := logger.WithFields(log.Fields{
        "command":   command,
        "directory": directory,
        "configFile": config,
    })

    return entry
}

