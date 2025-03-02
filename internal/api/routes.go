package api

import (
	"net/http"
)

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthCheckHandler)
	mux.HandleFunc("/video/upload", VideoUploadHandler)
	mux.HandleFunc("/video/", VideoStreamHandler)

	return mux
}
