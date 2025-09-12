package main

import (
	"log"

	"github.com/ramiroschettino/jwt-auth-api/internal/api"
	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"github.com/ramiroschettino/jwt-auth-api/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a la DB: ", err)
	}

	userRepo := repositories.NewUserRepository(db)
	noteRepo := repositories.NewNoteRepository(db)
	authService := services.NewAuthService(userRepo, cfg)
	noteService := services.NewNoteService(noteRepo)

	handler := api.NewAPIHandler(authService, noteService)
	router := api.NewRouter(handler)

	log.Println("Servidor escuchando en :8080")
	if err := api.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Error iniciando servidor: ", err)
	}
}
