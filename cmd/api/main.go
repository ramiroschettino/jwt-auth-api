package main

import (
	"fmt"
	"log"

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

	user, err := authService.Register("testuser2", "testpass", "user")
	if err != nil {
		log.Fatal("Error registrando usuario: ", err)
	}
	fmt.Println("Usuario registrado:", user.Username)

	token, err := authService.Login("testuser2", "testpass")
	if err != nil {
		log.Fatal("Error en login: ", err)
	}
	fmt.Println("Token JWT:", token)

	note, err := noteService.CreateNote("Segunda nota", "Otro contenido", user.ID)
	if err != nil {
		log.Fatal("Error creando nota: ", err)
	}
	fmt.Println("Nota creada:", note.Title)

	notes, err := noteService.GetNotesByUserID(user.ID)
	if err != nil {
		log.Fatal("Error listando notas: ", err)
	}
	for _, n := range notes {
		fmt.Printf("Nota: %s, Contenido: %s\n", n.Title, n.Content)
	}
}
