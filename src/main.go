package main

import (
	"os"

	"github.com/facecord/src/logger"
)

func main() {
	if os.Getenv("ENVIRONMENT") == "production" {
		logger.SetLevel(logger.InfoLevel)
	} else {
		logger.SetLevel(logger.DebugLevel)
	}
	proxy, err := NewProxyBot()
	if err != nil {
		logger.Error(NoTag, "could not start proxy")
	}
	proxy.Run()
}
