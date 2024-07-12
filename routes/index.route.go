package routes

import (
	"github.com/gorilla/mux"
	"tbrBackend/handlers"
)

func RegisterRoutes(router *mux.Router) {
	RegisterSoundRoutes(router)
	RegisterHealthRoutes(router)
}

func RegisterHealthRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/health", handlers.HealthCheck)
}
