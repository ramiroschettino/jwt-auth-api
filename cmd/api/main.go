package main

import (
	"fmt"
	"log"

	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	repository "github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a la DB: ", err)
	}

	userRepo := repository.NewUserRepository(db)
	noteRepo := repository.NewNoteRepository(db)

	user := &models.User{
		Username: "testuser",
		Password: "testpass",
		Role:     "user",
	}
	if err := userRepo.CreateUser(user); err != nil {
		log.Fatal("Error creando usuario: ", err)
	}
	fmt.Println("Usuario creado:", user.Username)

	note := &models.Note{
		Title:   "Mi primera nota",
		Content: "Contenido de prueba",
		UserID:  user.ID,
	}
	if err := noteRepo.CreateNote(note); err != nil {
		log.Fatal("Error creando nota: ", err)
	}
	fmt.Println("Nota creada:", note.Title)

	notes, err := noteRepo.FindNotesByUserID(user.ID)
	if err != nil {
		log.Fatal("Error buscando notas: ", err)
	}
	for _, n := range notes {
		fmt.Printf("Nota: %s, Contenido: %s\n", n.Title, n.Content)
	}
}
