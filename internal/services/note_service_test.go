package services

import (
	"testing"

	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNoteService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&models.User{}, &models.Note{})
	assert.NoError(t, err)

	user := &models.User{Username: "testuser", Password: "testpass", Role: "user"}
	err = db.Create(user).Error
	assert.NoError(t, err)

	noteRepo := repositories.NewNoteRepository(db)
	noteService := NewNoteService(noteRepo)

	t.Run("CreateNote", func(t *testing.T) {
		note, err := noteService.CreateNote("Test Note", "Test Content", user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, "Test Note", note.Title)
		assert.Equal(t, "Test Content", note.Content)
		assert.Equal(t, user.ID, note.UserID)

		var dbNote models.Note
		err = db.Where("user_id = ?", user.ID).First(&dbNote).Error
		assert.NoError(t, err)
		assert.Equal(t, note.ID, dbNote.ID)
	})

	t.Run("GetNotesByUserID", func(t *testing.T) {
		_, err := noteService.CreateNote("Another Note", "More Content", user.ID)
		assert.NoError(t, err)

		notes, err := noteService.GetNotesByUserID(user.ID)
		assert.NoError(t, err)
		assert.Len(t, notes, 2)
		assert.Equal(t, "Test Note", notes[0].Title)
		assert.Equal(t, "Another Note", notes[1].Title)
	})
}
