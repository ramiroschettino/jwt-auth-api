package main

import (
	"fmt"
	"log"

	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a la DB: ", err)
	}
	err = db.AutoMigrate(&models.User{}, &models.Note{})
	if err != nil {
		log.Fatal("Error migrando modelos: ", err)
	}
	fmt.Println("Tablas User y Note creadas en la base de datos")
}
