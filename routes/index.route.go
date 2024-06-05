package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("API is up and running haha"))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func RegisterRoutes(router *mux.Router) {
	RegisterSoundRoutes(router)
	RegisterHealthRoutes(router)
}

func RegisterHealthRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/health", HealthCheck)
}
