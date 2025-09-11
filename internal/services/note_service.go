package services

import (
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
)

type NoteService struct {
	noteRepo *repositories.NoteRepository
}

func NewNoteService(noteRepo *repositories.NoteRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo}
}

func (s *NoteService) CreateNote(title, content string, userID uint) (*models.Note, error) {
	note := &models.Note{
		Title:   title,
		Content: content,
		UserID:  userID,
	}
	if err := s.noteRepo.CreateNote(note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *NoteService) GetNotesByUserID(userID uint) ([]models.Note, error) {
	return s.noteRepo.FindNotesByUserID(userID)
}
