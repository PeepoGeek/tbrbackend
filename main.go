package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"tbrBackend/db"
	_ "tbrBackend/docs"
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
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	fmt.Println("Server is running on port 8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
		return
	}

}
