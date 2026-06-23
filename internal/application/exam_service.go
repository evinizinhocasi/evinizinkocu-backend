package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"evinizinkocu-backend/internal/domain"
)

type ExamService struct {
	examRepo    domain.TrialExamRepository
	catalogRepo domain.CatalogRepository
	studentRepo domain.StudentRepository
}

func NewExamService(
	examRepo domain.TrialExamRepository,
	catalogRepo domain.CatalogRepository,
	studentRepo domain.StudentRepository,
) *ExamService {
	return &ExamService{
		examRepo:    examRepo,
		catalogRepo: catalogRepo,
		studentRepo: studentRepo,
	}
}

type SubjectResultInput struct {
	SubjectID string `json:"subject_id"`
	Correct   int    `json:"correct"`
	Incorrect int    `json:"incorrect"`
	Blank     int    `json:"blank"`
}

func (s *ExamService) CreateTrialExam(
	ctx context.Context,
	studentID, creatorID, examName string,
	examDate time.Time,
	examTypeID string,
	score *float64,
	ranking *int,
	coachComment string,
	resultsInput []*SubjectResultInput,
) (*domain.TrialExam, error) {
	// 1. Verify student exists and is active
	stud, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return nil, domain.ErrStudentNotFound
	}
	if !stud.User.IsActive {
		return nil, domain.ErrPassiveAccount
	}

	// 2. Fetch Exam Type to get Divisor
	et, err := s.catalogRepo.GetExamTypeByID(ctx, examTypeID)
	if err != nil {
		return nil, errors.New("geçersiz sınav türü")
	}
	divisor := float64(et.Divisor)
	if divisor <= 0 {
		divisor = 4 // Default fallback
	}

	// 3. Process results and calculate nets
	var totalNet float64
	var results []*domain.TrialExamSubjectResult

	for _, input := range resultsInput {
		if input.Correct < 0 || input.Incorrect < 0 || input.Blank < 0 {
			return nil, errors.New("doğru, yanlış veya boş sayıları negatif olamaz")
		}

		// Calculate subject net
		net := float64(input.Correct) - (float64(input.Incorrect) / divisor)

		// Fetch subject info to check max question count
		subj, err := s.catalogRepo.GetSubjectByID(ctx, input.SubjectID)
		if err != nil {
			return nil, fmt.Errorf("geçersiz ders ID: %s", input.SubjectID)
		}
		if subj.QuestionCount != nil {
			totalQuestions := input.Correct + input.Incorrect + input.Blank
			if totalQuestions > *subj.QuestionCount {
				return nil, fmt.Errorf("%s dersi için girilen soru sayısı limitini (%d) aşıyor", subj.Name, *subj.QuestionCount)
			}
		}

		totalNet += net
		results = append(results, &domain.TrialExamSubjectResult{
			SubjectID: input.SubjectID,
			Correct:   input.Correct,
			Incorrect: input.Incorrect,
			Blank:     input.Blank,
			Net:       net,
		})
	}

	exam := &domain.TrialExam{
		StudentID:    studentID,
		CreatorID:    creatorID,
		ExamName:     examName,
		ExamDate:     examDate,
		ExamTypeID:   examTypeID,
		TotalNet:     totalNet,
		Score:        score,
		Ranking:      ranking,
		CoachComment: coachComment,
		Results:      results,
	}

	err = s.examRepo.CreateExam(ctx, exam)
	return exam, err
}

func (s *ExamService) UpdateTrialExam(
	ctx context.Context,
	id, editorID, examName string,
	examDate time.Time,
	score *float64,
	ranking *int,
	coachComment string,
	resultsInput []*SubjectResultInput,
) error {
	exam, err := s.examRepo.GetExamByID(ctx, id)
	if err != nil {
		return err
	}

	// editor permissions
	_, err = s.studentRepo.GetStudentByID(ctx, exam.StudentID)
	if err != nil {
		return err
	}

	// Fetch Exam Type to get Divisor
	et, err := s.catalogRepo.GetExamTypeByID(ctx, exam.ExamTypeID)
	if err != nil {
		return errors.New("geçersiz sınav türü")
	}
	divisor := float64(et.Divisor)

	// Process results
	var totalNet float64
	var results []*domain.TrialExamSubjectResult

	for _, input := range resultsInput {
		if input.Correct < 0 || input.Incorrect < 0 || input.Blank < 0 {
			return errors.New("doğru, yanlış veya boş sayıları negatif olamaz")
		}

		net := float64(input.Correct) - (float64(input.Incorrect) / divisor)

		subj, err := s.catalogRepo.GetSubjectByID(ctx, input.SubjectID)
		if err != nil {
			return fmt.Errorf("geçersiz ders ID: %s", input.SubjectID)
		}
		if subj.QuestionCount != nil {
			totalQuestions := input.Correct + input.Incorrect + input.Blank
			if totalQuestions > *subj.QuestionCount {
				return fmt.Errorf("%s dersi için girilen soru sayısı limitini (%d) aşıyor", subj.Name, *subj.QuestionCount)
			}
		}

		totalNet += net
		results = append(results, &domain.TrialExamSubjectResult{
			SubjectID: input.SubjectID,
			Correct:   input.Correct,
			Incorrect: input.Incorrect,
			Blank:     input.Blank,
			Net:       net,
		})
	}

	exam.ExamName = examName
	exam.ExamDate = examDate
	exam.TotalNet = totalNet
	exam.Score = score
	exam.Ranking = ranking
	exam.Results = results

	// Only coach can edit coachComment
	if editorID != exam.StudentID {
		// Editor is coach
		exam.CoachComment = coachComment
	} else {
		// Editor is student, keep previous comment
		// (Students cannot edit coach comment)
	}

	return s.examRepo.UpdateExam(ctx, exam)
}

func (s *ExamService) DeleteTrialExam(ctx context.Context, id string) error {
	return s.examRepo.DeleteExam(ctx, id)
}

func (s *ExamService) ListExamsByStudent(
	ctx context.Context,
	studentID string,
	examTypeID string,
	startDate, endDate *time.Time,
) ([]*domain.TrialExam, error) {
	return s.examRepo.ListExamsByStudent(ctx, studentID, examTypeID, startDate, endDate)
}

func (s *ExamService) GetTrialExamByID(ctx context.Context, id string) (*domain.TrialExam, error) {
	return s.examRepo.GetExamByID(ctx, id)
}

// Custom format helper
