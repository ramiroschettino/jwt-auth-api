package services

import (
	"testing"

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
	s.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.NoError(err)

	err = s.db.AutoMigrate(&models.User{}, &models.Note{})
	s.NoError(err)

	s.testUser = &models.User{
		Username: "testuser",
		Password: "hashedpass",
		Role:     "user",
	}
	err = s.db.Create(s.testUser).Error
	s.NoError(err)

	noteRepo := repositories.NewNoteRepository(s.db)
	s.noteService = NewNoteService(noteRepo)
}

func TestNoteService(t *testing.T) {
	suite.Run(t, new(NoteServiceTestSuite))
}

func (s *NoteServiceTestSuite) TestNoteService() {
	t := s.T()

	t.Run("CreateNote", func(t *testing.T) {
		note, err := s.noteService.CreateNote("Test Note", "Test Content", s.testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, "Test Note", note.Title)
		assert.Equal(t, "Test Content", note.Content)
		assert.Equal(t, s.testUser.ID, note.UserID)

		var dbNote models.Note
		err = s.db.Where("user_id = ?", s.testUser.ID).First(&dbNote).Error
		assert.NoError(t, err)
		assert.Equal(t, note.ID, dbNote.ID)
	})

	t.Run("GetNotesByUserID", func(t *testing.T) {
		_, err := s.noteService.CreateNote("Another Note", "More Content", s.testUser.ID)
		assert.NoError(t, err)

		notes, err := s.noteService.GetNotesByUserID(s.testUser.ID)
		assert.NoError(t, err)
		assert.Len(t, notes, 2)
		assert.Equal(t, "Test Note", notes[0].Title)
		assert.Equal(t, "Another Note", notes[1].Title)
	})
}
