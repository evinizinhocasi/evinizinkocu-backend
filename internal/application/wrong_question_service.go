package application

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/storage"

	"github.com/google/uuid"
)

var (
	ErrWrongQuestionNotFound = errors.New("yanlış soru bulunamadı")
	ErrUnauthorizedAction    = errors.New("bu işlem için yetkiniz yok")
)

type WrongQuestionService struct {
	repo        domain.WrongQuestionRepository
	studentRepo domain.StudentRepository
	r2Storage   *storage.R2Storage
}

func NewWrongQuestionService(
	repo domain.WrongQuestionRepository,
	studentRepo domain.StudentRepository,
	r2Storage *storage.R2Storage,
) *WrongQuestionService {
	return &WrongQuestionService{
		repo:        repo,
		studentRepo: studentRepo,
		r2Storage:   r2Storage,
	}
}

func (s *WrongQuestionService) Create(
	ctx context.Context,
	studentID uuid.UUID,
	creatorUserID uuid.UUID,
	creatorRole string,
	subjectID *uuid.UUID,
	title string,
	note string,
	fileName string,
	fileBytes []byte,
	contentType string,
) (*domain.WrongQuestion, error) {
	// Security / Role check
	if creatorRole == "student" && creatorUserID != studentID {
		return nil, ErrUnauthorizedAction
	}

	if creatorRole == "coach" {
		student, err := s.studentRepo.GetStudentByID(ctx, studentID.String())
		if err != nil || student == nil || student.CoachID != creatorUserID.String() {
			return nil, ErrUnauthorizedAction
		}
	}

	// Upload to Cloudflare R2
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = ".jpg"
	}

	r2FileName := fmt.Sprintf("wq_%s_%d_%d%s", studentID.String()[:8], time.Now().UnixNano(), rand.Intn(10000), ext)
	imageURL, err := s.r2Storage.UploadFile(r2FileName, fileBytes, contentType)
	if err != nil {
		return nil, fmt.Errorf("görsel yükleme hatası: %w", err)
	}

	input := domain.CreateWrongQuestionInput{
		StudentID:     studentID,
		CreatorUserID: creatorUserID,
		SubjectID:     subjectID,
		ImageURL:      imageURL,
		Title:         title,
		Note:          note,
	}

	return s.repo.Create(ctx, input)
}

func (s *WrongQuestionService) ListByStudent(
	ctx context.Context,
	studentID uuid.UUID,
	subjectID *uuid.UUID,
	requestingUserID uuid.UUID,
	requestingUserRole string,
) ([]domain.WrongQuestion, error) {
	// Security check
	if requestingUserRole == "student" && requestingUserID != studentID {
		return nil, ErrUnauthorizedAction
	}

	if requestingUserRole == "coach" {
		student, err := s.studentRepo.GetStudentByID(ctx, studentID.String())
		if err != nil || student == nil || student.CoachID != requestingUserID.String() {
			return nil, ErrUnauthorizedAction
		}
	}

	return s.repo.ListByStudent(ctx, studentID, subjectID)
}

func (s *WrongQuestionService) Update(
	ctx context.Context,
	input domain.UpdateWrongQuestionInput,
	requestingUserID uuid.UUID,
	requestingUserRole string,
) error {
	existing, err := s.repo.GetByID(ctx, input.ID)
	if err != nil || existing == nil {
		return ErrWrongQuestionNotFound
	}

	if requestingUserRole == "student" && requestingUserID != existing.StudentID {
		return ErrUnauthorizedAction
	}

	if requestingUserRole == "coach" {
		student, err := s.studentRepo.GetStudentByID(ctx, existing.StudentID.String())
		if err != nil || student == nil || student.CoachID != requestingUserID.String() {
			return ErrUnauthorizedAction
		}
	}

	return s.repo.Update(ctx, input)
}

func (s *WrongQuestionService) ToggleResolved(
	ctx context.Context,
	id uuid.UUID,
	isResolved bool,
	requestingUserID uuid.UUID,
	requestingUserRole string,
) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return ErrWrongQuestionNotFound
	}

	if requestingUserRole == "student" && requestingUserID != existing.StudentID {
		return ErrUnauthorizedAction
	}

	if requestingUserRole == "coach" {
		student, err := s.studentRepo.GetStudentByID(ctx, existing.StudentID.String())
		if err != nil || student == nil || student.CoachID != requestingUserID.String() {
			return ErrUnauthorizedAction
		}
	}

	return s.repo.ToggleResolved(ctx, id, isResolved)
}

func (s *WrongQuestionService) Delete(
	ctx context.Context,
	id uuid.UUID,
	requestingUserID uuid.UUID,
	requestingUserRole string,
) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return ErrWrongQuestionNotFound
	}

	if requestingUserRole == "student" && requestingUserID != existing.StudentID {
		return ErrUnauthorizedAction
	}

	if requestingUserRole == "coach" {
		student, err := s.studentRepo.GetStudentByID(ctx, existing.StudentID.String())
		if err != nil || student == nil || student.CoachID != requestingUserID.String() {
			return ErrUnauthorizedAction
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *WrongQuestionService) DownloadMedia(fileName string) ([]byte, string, error) {
	return s.r2Storage.DownloadFile(fileName)
}
