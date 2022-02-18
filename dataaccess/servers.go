package dataaccess

import (
	"errors"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enroller/dataaccess/fileio"
)


func EnrollServer(server objects.Server) error {
	slog.PrintInfo("Enrolling server", server.ServerID)
	slog.PrintTrace(server)

	switch Backend {
		case "file":
			return fileio.EnrollServer(server)
	}

	return errors.New("Selected backend is unknown")
}
