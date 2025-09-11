package repositories

import (
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) CreateNote(note *models.Note) error {
	return r.db.Create(note).Error
}

func (r *NoteRepository) FindNotesByUserID(userID uint) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Where("user_id = ?", userID).Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}
