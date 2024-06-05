package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var DSN = "host=localhost user=joseSuperior password=a12345678 dbname=tbrdb port=5432"

// DB Se declara la variable afuerade la función para que sea accesible desde cualquier parte del paquete
var DB *gorm.DB

func ConnectDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Database connected")
	}

}
