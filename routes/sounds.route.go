package routes

import (
	"github.com/gorilla/mux"
	"tbrBackend/handlers"
)

func RegisterSoundRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/sound", handlers.GetSoundsHandler).Methods("GET")
	router.HandleFunc("/api/v1/sound/{id}", handlers.GetSoundHandler).Methods("GET")
	router.HandleFunc("/api/v1/sound", handlers.CreateSoundHandler).Methods("POST")
	router.HandleFunc("/api/v1/sound/{id}", handlers.UpdateSoundHandler).Methods("PUT")
	router.HandleFunc("/api/v1/sound/{id}", handlers.DeleteSoundHandler).Methods("DELETE")
}
