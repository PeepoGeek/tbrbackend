package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"tbrBackend/db"
	"tbrBackend/models"
	"tbrBackend/routes"
	"tbrBackend/services"
)

func main() {
	db.ConnectDB()
	s3Client := services.NewS3Client()
	s3Client.ListBuckets()

	erro := db.DB.AutoMigrate(models.Sound{})
	if erro != nil {
		log.Fatal(erro)
		return
	}
	router := mux.NewRouter()
	routes.RegisterRoutes(router)
	fmt.Println("Server is running on port 8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
		return
	}
}
