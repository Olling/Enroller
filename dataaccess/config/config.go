package config

import (
	"os"
	"time"
	"encoding/json"
	"github.com/Olling/slog"
)

type configuration struct {
	FileBackendDirectory   	string
	ScriptDirectory        	string
	EnrollmentScriptPath   	string
	NotificationScriptPath 	string
	TempPath               	string
	LogPath                	string
	Port                   	string
	DefaultInventoryName   	string
	ServerIDRegexp         	string
	EnrolldURL		string

	TlsPort     		string
	TlsCert			string
	TlsKey			string

	Timeout     		int
	Interval		time.Duration
	ThreadNumber		int
	WorkerID		string
}

var (
	Configuration configuration
)

func InitializeConfiguration(path string) {
	file, _ := os.Open(path)

	err := json.NewDecoder(file).Decode(&Configuration)
	if err != nil {
		slog.PrintError("Error while reading the configuration file - Exiting")
		slog.PrintError(err)
		os.Exit(1)
	}

	_, err = os.Stat(Configuration.FileBackendDirectory)

	if os.IsNotExist(err) {
		err := os.MkdirAll(Configuration.FileBackendDirectory, 0744)
		if err != nil {
			slog.PrintError(err)
		} else {
			slog.PrintInfo("Created: " + Configuration.FileBackendDirectory)
		}
	}

}
