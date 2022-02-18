package dataaccess

import (
	"fmt"
	"sync"
	"errors"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/fileio"
)

var (
	Backend			string
	Users			map[string]objects.User
	Scripts			map[string]objects.Script
	SyncGetInventoryMutex	sync.Mutex
)

func Initialize(backend string) {
	Backend = backend
}

func InitializeAuthentication() {
	slog.PrintDebug("Initializing Authentication")
	err := LoadAuthentication()
	if err != nil {
		slog.PrintError("Failed to load authentication", err)
	}
}

func LoadAuthentication() error {
	switch Backend {
		case "file":
			return fileio.LoadFromFile(&Users, "/etc/enroller/auth.json")
	}
	return errors.New("Selected backend is unknown")
}

func RunScript(scriptPath string, server objects.Server, scriptID string, timeout int) error {
	if server.ServerID == "" {
		slog.PrintError("Failed to call", scriptID, "script - ServerID is empty!")
		return fmt.Errorf("ServerID was not given")
	}

	err := fileio.RunScript(scriptPath, server, scriptID, timeout)

	return err
}

func LoadScripts() error {
	slog.PrintDebug("Loading Scripts")
	return fileio.LoadScripts(Scripts)
}

func CheckScriptPath(path string) error {
	return fileio.CheckScriptPath(path)
}
