package application

import (
	"context"
	"fmt"
	"time"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

var turkishMonths = []string{
	"Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
	"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık",
}

type MonthlyReportService struct {
	pool        *pgxpool.Pool
	studentRepo domain.StudentRepository
}

func NewMonthlyReportService(pool *pgxpool.Pool, studentRepo domain.StudentRepository) *MonthlyReportService {
	return &MonthlyReportService{
		pool:        pool,
		studentRepo: studentRepo,
	}
}

func (s *MonthlyReportService) GetStudentMonthlyReport(
	ctx context.Context,
	studentID string,
	year, month int,
) (*domain.MonthlyReportData, error) {
	if year <= 2000 {
		year = time.Now().Year()
	}
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	monthName := fmt.Sprintf("%s %d", turkishMonths[month-1], year)

	// 1. Student Details
	student, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return nil, domain.ErrStudentNotFound
	}

	studentName := ""
	if student.User != nil {
		studentName = fmt.Sprintf("%s %s", student.User.FirstName, student.User.LastName)
	}

	report := &domain.MonthlyReportData{
		StudentID:         student.ID,
		StudentName:       studentName,
		ClassLevel:        student.ClassLevel,
		StudyTrack:        student.StudyTrack,
		ExamTypeName:      student.ExamTypeName,
		CoachName:         student.CoachName,
		TargetSchoolName:  student.TargetSchoolName,
		TargetSchoolPhoto: student.TargetSchoolPhoto,
		Year:              year,
		Month:             month,
		MonthName:         monthName,
		Summary:           domain.MonthlyReportSummary{},
		SubjectBreakdown:  []domain.SubjectReportDetail{},
		TrialExams:        []domain.TrialExamReportDetail{},
	}

	// 2. Question Solving Breakdown by Subject
	qQuery := `
		SELECT
			COALESCE(sub.name, 'Genel'),
			COALESCE(SUM(q.correct_count + q.wrong_count + q.empty_count), 0)::int,
			COALESCE(SUM(q.correct_count), 0)::int,
			COALESCE(SUM(q.wrong_count), 0)::int,
			COALESCE(SUM(q.empty_count), 0)::int,
			COALESCE(SUM(q.net_count), 0)::float8
		FROM question_solving_entries q
		LEFT JOIN subjects sub ON q.subject_id = sub.id
		WHERE q.student_id = $1 AND q.date >= $2 AND q.date < $3
		GROUP BY sub.name
		ORDER BY sub.name ASC
	`
	qRows, err := s.pool.Query(ctx, qQuery, studentID, startDate, endDate)
	if err == nil {
		defer qRows.Close()
		for qRows.Next() {
			var detail domain.SubjectReportDetail
			if scanErr := qRows.Scan(
				&detail.SubjectName,
				&detail.QuestionsSolved,
				&detail.Correct,
				&detail.Wrong,
				&detail.Empty,
				&detail.Net,
			); scanErr == nil {
				report.SubjectBreakdown = append(report.SubjectBreakdown, detail)
				report.Summary.TotalQuestionsSolved += detail.QuestionsSolved
				report.Summary.TotalCorrect += detail.Correct
				report.Summary.TotalWrong += detail.Wrong
				report.Summary.TotalEmpty += detail.Empty
				report.Summary.TotalNet += detail.Net
			}
		}
	}

	// 3. Homework / Tasks
	hwQuery := `
		SELECT
			COUNT(*)::int,
			COALESCE(COUNT(*) FILTER (WHERE status = 'completed'), 0)::int
		FROM homework
		WHERE student_id = $1 AND created_at >= $2 AND created_at < $3
	`
	_ = s.pool.QueryRow(ctx, hwQuery, studentID, startDate, endDate).Scan(
		&report.Summary.TotalHomeworkAssigned,
		&report.Summary.TotalHomeworkCompleted,
	)
	if report.Summary.TotalHomeworkAssigned > 0 {
		report.Summary.HomeworkCompletionRate = float64(report.Summary.TotalHomeworkCompleted) / float64(report.Summary.TotalHomeworkAssigned) * 100.0
	}

	// 4. Trial Exams
	teQuery := `
		SELECT
			title,
			exam_date::text,
			total_net,
			score,
			ranking
		FROM trial_exams
		WHERE student_id = $1 AND exam_date >= $2 AND exam_date < $3
		ORDER BY exam_date ASC
	`
	teRows, err := s.pool.Query(ctx, teQuery, studentID, startDate, endDate)
	if err == nil {
		defer teRows.Close()
		totalNetSum := 0.0
		for teRows.Next() {
			var te domain.TrialExamReportDetail
			if scanErr := teRows.Scan(&te.ExamTitle, &te.ExamDate, &te.TotalNet, &te.Score, &te.Rank); scanErr == nil {
				report.TrialExams = append(report.TrialExams, te)
				totalNetSum += te.TotalNet
			}
		}
		report.Summary.TrialExamsCount = len(report.TrialExams)
		if report.Summary.TrialExamsCount > 0 {
			report.Summary.AverageTrialNet = totalNetSum / float64(report.Summary.TrialExamsCount)
		}
	}

	// 5. Wrong Questions Logged & Resolved
	wqQuery := `
		SELECT
			COUNT(*)::int,
			COALESCE(COUNT(*) FILTER (WHERE is_resolved = true), 0)::int
		FROM wrong_questions
		WHERE student_id = $1 AND created_at >= $2 AND created_at < $3
	`
	_ = s.pool.QueryRow(ctx, wqQuery, studentID, startDate, endDate).Scan(
		&report.Summary.WrongQuestionsLogged,
		&report.Summary.WrongQuestionsResolved,
	)

	// 6. Meetings / Coaching Sessions
	meetQuery := `
		SELECT COUNT(*)::int
		FROM meetings
		WHERE student_id = $1 AND meeting_date >= $2 AND meeting_date < $3
	`
	_ = s.pool.QueryRow(ctx, meetQuery, studentID, startDate, endDate).Scan(
		&report.Summary.MeetingsCount,
	)

	return report, nil
}
