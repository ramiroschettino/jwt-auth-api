package services

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type NoteServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	noteService *NoteService
	testUser    *models.User
}

func (s *NoteServiceTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.NoError(err)

	// Migrar tablas
	s.NoError(s.db.AutoMigrate(&models.User{}, &models.Note{}))

	// Limpiar datos
	s.db.Exec("DELETE FROM notes")
	s.db.Exec("DELETE FROM users")

	// Crear usuario de prueba
	s.testUser = &models.User{
		Username: "testuser",
		Password: "hashedpass",
		Role:     "admin",
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	err = s.db.Create(s.testUser).Error
	s.NoError(err)

	noteRepo := repositories.NewNoteRepository(s.db)
	s.noteService = NewNoteService(noteRepo)
}

func TestNoteService(t *testing.T) {
	suite.Run(t, new(NoteServiceTestSuite))
}

func (s *NoteServiceTestSuite) TestNoteOperations() {
	t := s.T()

	t.Run("Create Note", func(t *testing.T) {
		// Crear nota válida
		note, err := s.noteService.CreateNote("Test Note", "Test Content", s.testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, "Test Note", note.Title)
		assert.Equal(t, "Test Content", note.Content)
		assert.Equal(t, s.testUser.ID, note.UserID)

		// Verificar persistencia
		var dbNote models.Note
		err = s.db.Where("user_id = ?", s.testUser.ID).First(&dbNote).Error
		assert.NoError(t, err)
		assert.Equal(t, note.ID, dbNote.ID)
	})

	t.Run("Get Notes", func(t *testing.T) {
		// Crear varias notas
		_, err := s.noteService.CreateNote("Note 1", "Content 1", s.testUser.ID)
		assert.NoError(t, err)
		_, err = s.noteService.CreateNote("Note 2", "Content 2", s.testUser.ID)
		assert.NoError(t, err)

		// Obtener notas
		notes, err := s.noteService.GetNotesByUserID(s.testUser.ID)
		assert.NoError(t, err)
		assert.Len(t, notes, 3) // 2 nuevas + 1 del test anterior
		assert.Equal(t, "Test Note", notes[0].Title)
		assert.Equal(t, "Note 1", notes[1].Title)
		assert.Equal(t, "Note 2", notes[2].Title)
	})

	t.Run("Get Notes Empty", func(t *testing.T) {
		// Limpiar notas
		s.db.Exec("DELETE FROM notes")

		// Verificar que retorna slice vacío
		notes, err := s.noteService.GetNotesByUserID(s.testUser.ID)
		assert.NoError(t, err)
		assert.Empty(t, notes)
	})

	t.Run("Get Notes Non-Existent User", func(t *testing.T) {
		notes, err := s.noteService.GetNotesByUserID(999)
		assert.NoError(t, err)
		assert.Empty(t, notes)
	})
}
