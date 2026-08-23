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
	ErrTargetSchoolNotFound = errors.New("hedef okul bulunamadı")
	ErrSchoolNameRequired   = errors.New("okul adı zorunludur")
)

type TargetSchoolService struct {
	repo        domain.TargetSchoolRepository
	studentRepo domain.StudentRepository
	r2Storage   *storage.R2Storage
}

func NewTargetSchoolService(
	repo domain.TargetSchoolRepository,
	studentRepo domain.StudentRepository,
	r2Storage *storage.R2Storage,
) *TargetSchoolService {
	return &TargetSchoolService{
		repo:        repo,
		studentRepo: studentRepo,
		r2Storage:   r2Storage,
	}
}

func (s *TargetSchoolService) List(ctx context.Context, schoolType string, search string) ([]domain.TargetSchool, error) {
	return s.repo.List(ctx, schoolType, search)
}

func (s *TargetSchoolService) GetByID(ctx context.Context, id uuid.UUID) (*domain.TargetSchool, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TargetSchoolService) Create(
	ctx context.Context,
	name string,
	schoolType domain.TargetSchoolType,
	city string,
	minScore float64,
	percentile float64,
	ranking int,
	department string,
	fileName string,
	fileBytes []byte,
	contentType string,
) (*domain.TargetSchool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrSchoolNameRequired
	}

	var photoURL string
	if len(fileBytes) > 0 {
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext == "" {
			ext = ".jpg"
		}
		r2FileName := fmt.Sprintf("school_%d_%d%s", time.Now().UnixNano(), rand.Intn(10000), ext)
		uploadedURL, err := s.r2Storage.UploadFile(r2FileName, fileBytes, contentType)
		if err == nil {
			photoURL = uploadedURL
		}
	}

	input := domain.CreateTargetSchoolInput{
		Name:       name,
		Type:       schoolType,
		City:       city,
		PhotoURL:   photoURL,
		MinScore:   minScore,
		Percentile: percentile,
		Ranking:    ranking,
		Department: department,
	}

	return s.repo.Create(ctx, input)
}

func (s *TargetSchoolService) Update(
	ctx context.Context,
	id uuid.UUID,
	name string,
	schoolType domain.TargetSchoolType,
	city string,
	minScore float64,
	percentile float64,
	ranking int,
	department string,
	fileName string,
	fileBytes []byte,
	contentType string,
) (*domain.TargetSchool, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return nil, ErrTargetSchoolNotFound
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrSchoolNameRequired
	}

	photoURL := existing.PhotoURL
	if len(fileBytes) > 0 {
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext == "" {
			ext = ".jpg"
		}
		r2FileName := fmt.Sprintf("school_%d_%d%s", time.Now().UnixNano(), rand.Intn(10000), ext)
		uploadedURL, err := s.r2Storage.UploadFile(r2FileName, fileBytes, contentType)
		if err == nil {
			photoURL = uploadedURL
		}
	}

	input := domain.UpdateTargetSchoolInput{
		ID:         id,
		Name:       name,
		Type:       schoolType,
		City:       city,
		PhotoURL:   photoURL,
		MinScore:   minScore,
		Percentile: percentile,
		Ranking:    ranking,
		Department: department,
	}

	return s.repo.Update(ctx, input)
}

func (s *TargetSchoolService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *TargetSchoolService) UpdateStudentTargetSchool(
	ctx context.Context,
	studentID string,
	requestingUserID string,
	requestingUserRole string,
	targetSchoolID *string,
	targetSchoolName string,
	targetSchoolPhoto string,
	targetSchoolType string,
	fileName string,
	fileBytes []byte,
	contentType string,
) error {
	student, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil || student == nil {
		return domain.ErrStudentNotFound
	}

	// Security check: coach can only update their own students, student can only update themselves, admin can update anyone
	if requestingUserRole == domain.RoleCoach && student.CoachID != requestingUserID {
		return domain.ErrForbidden
	}
	if requestingUserRole == domain.RoleStudent && student.ID != requestingUserID {
		return domain.ErrForbidden
	}

	// If a custom photo is uploaded
	if len(fileBytes) > 0 {
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext == "" {
			ext = ".jpg"
		}
		r2FileName := fmt.Sprintf("student_target_%s_%d%s", studentID[:8], time.Now().UnixNano(), ext)
		uploadedURL, err := s.r2Storage.UploadFile(r2FileName, fileBytes, contentType)
		if err == nil {
			targetSchoolPhoto = uploadedURL
		}
	}

	// If linked to catalog school and no photo supplied, copy catalog photo
	if targetSchoolID != nil && *targetSchoolID != "" && targetSchoolPhoto == "" {
		schoolUUID, err := uuid.Parse(*targetSchoolID)
		if err == nil {
			catalogSchool, _ := s.repo.GetByID(ctx, schoolUUID)
			if catalogSchool != nil {
				targetSchoolPhoto = catalogSchool.PhotoURL
				if targetSchoolName == "" {
					targetSchoolName = catalogSchool.Name
				}
				if targetSchoolType == "" {
					targetSchoolType = string(catalogSchool.Type)
				}
			}
		}
	}

	return s.studentRepo.UpdateTargetSchool(
		ctx,
		studentID,
		targetSchoolID,
		strings.TrimSpace(targetSchoolName),
		strings.TrimSpace(targetSchoolPhoto),
		strings.TrimSpace(targetSchoolType),
	)
}
