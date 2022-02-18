package main

import (
	"os"
	"github.com/Olling/slog"
	"github.com/Olling/Enroller/api"
	"github.com/Olling/Enroller/dataaccess"
	"github.com/Olling/Enroller/worker"
	"github.com/Olling/Enroller/dataaccess/config"
)

func main() {
	dataaccess.Initialize("file")
	config.InitializeConfiguration("/etc/enroller/enroller.conf")

	dataaccess.LoadScripts()

	slog.SetLogLevel(slog.Trace)

	scriptPathErr := dataaccess.CheckScriptPath(config.Configuration.EnrollmentScriptPath)
	if scriptPathErr != nil {
		slog.PrintFatal("EnrollmentScriptPath Problem - stopping")
		os.Exit(1)
	}
	
	worker.Initialize()

	dataaccess.InitializeAuthentication()
	api.SetupRouter()
}
