package api

import (
	"os"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/gorilla/handlers"
	"github.com/Olling/Enroller/dataaccess/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter initializes the API routes
func SetupRouter() {
	router := mux.NewRouter()

	router.HandleFunc("/status", getStatus).Methods("GET")

	// add prometheus handler
	router.Handle("/metrics", promhttp.Handler())

	// enable logging
	loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)

	slog.PrintInfo("Listening on port: " + config.Configuration.Port + " (http) and port: " + config.Configuration.TlsPort + " (https)")

	go http.ListenAndServe(":"+config.Configuration.Port, loggedRouter)
	err := http.ListenAndServeTLS(":"+config.Configuration.TlsPort, config.Configuration.TlsCert, config.Configuration.TlsKey, loggedRouter)
	if err != nil {
		slog.PrintError("Error starting TLS: ", err)
	}
}


func rootHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Welcome to Enroller")
}
