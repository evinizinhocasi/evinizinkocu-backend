package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/repository"
)

type TrackersService struct {
	repo         *repository.PostgresTrackersRepository
	studentRepo  domain.StudentRepository
	examRepo     domain.TrialExamRepository
	catalogRepo  domain.CatalogRepository
}

func NewTrackersService(
	repo *repository.PostgresTrackersRepository,
	studentRepo domain.StudentRepository,
	examRepo domain.TrialExamRepository,
	catalogRepo domain.CatalogRepository,
) *TrackersService {
	return &TrackersService{
		repo:        repo,
		studentRepo: studentRepo,
		examRepo:    examRepo,
		catalogRepo: catalogRepo,
	}
}

func (s *TrackersService) resolveTopicID(ctx context.Context, subjectID string, topicNameOrID *string) (*string, error) {
	if topicNameOrID == nil || *topicNameOrID == "" {
		return nil, nil
	}

	nameOrID := *topicNameOrID

	// First check if it is a valid UUID by checking format and existence
	if len(nameOrID) == 36 && strings.Contains(nameOrID, "-") {
		topic, err := s.catalogRepo.GetTopicByID(ctx, nameOrID)
		if err == nil && topic != nil {
			return &topic.ID, nil
		}
	}

	// Otherwise, find or create topic by name
	topics, err := s.catalogRepo.ListTopicsBySubject(ctx, subjectID)
	if err == nil {
		for _, t := range topics {
			if strings.EqualFold(strings.TrimSpace(t.Name), strings.TrimSpace(nameOrID)) {
				return &t.ID, nil
			}
		}
	}

	newTopic := &domain.Topic{
		SubjectID: subjectID,
		Name:      strings.TrimSpace(nameOrID),
	}
	err = s.catalogRepo.CreateTopic(ctx, newTopic)
	if err != nil {
		return nil, err
	}

	return &newTopic.ID, nil
}

// --- Question Solving Tracker ---

func (s *TrackersService) CreateQuestionSolving(
	ctx context.Context,
	studentID, creatorID string,
	date time.Time,
	subjectID string,
	topicID *string,
	correct, incorrect, blank int,
	note string,
) (*domain.QuestionSolvingEntry, error) {
	if correct < 0 || incorrect < 0 || blank < 0 {
		return nil, errors.New("sayılar negatif olamaz")
	}

	resolvedTopicID, err := s.resolveTopicID(ctx, subjectID, topicID)
	if err != nil {
		return nil, err
	}

	// Calculate net: correct - incorrect/divisor. Let's find exam type from student's profile
	student, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return nil, err
	}

	et, err := s.catalogRepo.GetExamTypeByID(ctx, student.ExamTypeID)
	if err != nil {
		return nil, errors.New("ders programı veya sınav türü hatalı")
	}

	net := float64(correct) - (float64(incorrect) / float64(et.Divisor))

	entry := &domain.QuestionSolvingEntry{
		StudentID: studentID,
		CreatorID: creatorID,
		Date:      date,
		SubjectID: subjectID,
		TopicID:   resolvedTopicID,
		Correct:   correct,
		Incorrect: incorrect,
		Blank:     blank,
		Net:       net,
		Note:      note,
	}

	err = s.repo.CreateQuestionSolving(ctx, entry)
	return entry, err
}

func (s *TrackersService) GetQuestionSolvingByID(ctx context.Context, id string) (*domain.QuestionSolvingEntry, error) {
	return s.repo.GetQuestionSolvingByID(ctx, id)
}

func (s *TrackersService) UpdateQuestionSolving(
	ctx context.Context,
	id string,
	date time.Time,
	subjectID string,
	topicID *string,
	correct, incorrect, blank int,
	note string,
) error {
	e, err := s.repo.GetQuestionSolvingByID(ctx, id)
	if err != nil {
		return err
	}

	resolvedTopicID, err := s.resolveTopicID(ctx, subjectID, topicID)
	if err != nil {
		return err
	}

	student, err := s.studentRepo.GetStudentByID(ctx, e.StudentID)
	if err != nil {
		return err
	}

	et, err := s.catalogRepo.GetExamTypeByID(ctx, student.ExamTypeID)
	if err != nil {
		return err
	}

	net := float64(correct) - (float64(incorrect) / float64(et.Divisor))

	e.Date = date
	e.SubjectID = subjectID
	e.TopicID = resolvedTopicID
	e.Correct = correct
	e.Incorrect = incorrect
	e.Blank = blank
	e.Net = net
	e.Note = note

	return s.repo.UpdateQuestionSolving(ctx, e)
}

func (s *TrackersService) DeleteQuestionSolving(ctx context.Context, id string) error {
	return s.repo.DeleteQuestionSolving(ctx, id)
}

func (s *TrackersService) ListQuestionSolvingByStudent(ctx context.Context, studentID string, startDate, endDate *time.Time) ([]*domain.QuestionSolvingEntry, error) {
	return s.repo.ListQuestionSolvingByStudent(ctx, studentID, startDate, endDate)
}

// --- Homework Tracker ---

func (s *TrackersService) CreateHomework(
	ctx context.Context,
	studentID, coachID, title, description, subjectID string,
	topicID *string,
	source, pageRange, url string,
	start, due time.Time,
) (*domain.Homework, error) {
	resolvedTopicID, err := s.resolveTopicID(ctx, subjectID, topicID)
	if err != nil {
		return nil, err
	}

	hw := &domain.Homework{
		StudentID:      studentID,
		CreatorCoachID: coachID,
		Title:          title,
		Description:    description,
		SubjectID:      subjectID,
		TopicID:        resolvedTopicID,
		Source:         source,
		PageRange:      pageRange,
		URL:            url,
		StartDate:      start,
		DueDate:        due,
		Status:         "waiting",
	}

	err = s.repo.CreateHomework(ctx, hw)
	return hw, err
}

func (s *TrackersService) UpdateHomeworkStatus(ctx context.Context, id, updaterID, updaterRole, newStatus, explanation string) error {
	hw, err := s.repo.GetHomeworkByID(ctx, id)
	if err != nil {
		return err
	}

	// State transitions
	switch newStatus {
	case "started":
		if updaterRole != domain.RoleStudent || hw.StudentID != updaterID {
			return errors.New("ödevi sadece ödev sahibi öğrenci başlatabilir")
		}
		if hw.Status != "waiting" && hw.Status != "started" {
			return errors.New("geçersiz durum geçişi")
		}
		hw.Status = "started"

	case "awaiting_approval":
		if updaterRole != domain.RoleStudent || hw.StudentID != updaterID {
			return errors.New("ödevi sadece ödev sahibi onay talebine gönderebilir")
		}
		if hw.Status != "started" {
			return errors.New("ödev başlatılmadan onay talebine gönderilemez")
		}
		hw.Status = "awaiting_approval"

	case "completed":
		if updaterRole != domain.RoleCoach && updaterRole != domain.RoleSuperadmin {
			return errors.New("sadece koç ödevi onaylayabilir")
		}
		hw.Status = "completed"
		hw.CoachExplanation = "" // Clear reject explanation

	case "started_reject": // Rejecting back to started
		if updaterRole != domain.RoleCoach && updaterRole != domain.RoleSuperadmin {
			return errors.New("sadece koç ödevi reddedebilir")
		}
		if hw.Status != "awaiting_approval" {
			return errors.New("yalnızca onay bekleyen ödevler reddedilebilir")
		}
		hw.Status = "started"
		hw.CoachExplanation = explanation

	default:
		return errors.New("geçersiz durum")
	}

	return s.repo.UpdateHomework(ctx, hw)
}

func (s *TrackersService) ListHomeworkByStudent(ctx context.Context, studentID string) ([]*domain.Homework, error) {
	return s.repo.ListHomeworkByStudent(ctx, studentID)
}

func (s *TrackersService) DeleteHomework(ctx context.Context, id string) error {
	return s.repo.DeleteHomework(ctx, id)
}

// --- Student Resources ---

func (s *TrackersService) CreateResource(
	ctx context.Context,
	studentID, name, publisher, subjectID, description string,
	totalPages, completedPages int,
	status string,
) (*domain.Resource, error) {
	if totalPages <= 0 {
		return nil, errors.New("toplam sayfa sayısı 0 veya negatif olamaz")
	}
	if completedPages < 0 || completedPages > totalPages {
		return nil, errors.New("tamamlanan sayfa sayısı geçersiz")
	}

	progress := int((float64(completedPages) / float64(totalPages)) * 100)

	res := &domain.Resource{
		StudentID:          studentID,
		Name:               name,
		Publisher:          publisher,
		SubjectID:          subjectID,
		Description:        description,
		TotalPages:         totalPages,
		CompletedPages:     completedPages,
		ProgressPercentage: progress,
		Status:             status,
	}

	err := s.repo.CreateResource(ctx, res)
	return res, err
}

func (s *TrackersService) UpdateResourceProgress(ctx context.Context, id string, completedPages, progressPercentage int, status string) error {
	res, err := s.repo.GetResourceByID(ctx, id)
	if err != nil {
		return err
	}

	if completedPages < 0 || completedPages > res.TotalPages {
		return errors.New("tamamlanan sayfa sayısı geçersiz")
	}

	// Auto-update progress percentage if completed pages is changed
	if completedPages != res.CompletedPages {
		res.CompletedPages = completedPages
		res.ProgressPercentage = int((float64(completedPages) / float64(res.TotalPages)) * 100)
	} else {
		// If page is same, allow manual override percentage
		if progressPercentage >= 0 && progressPercentage <= 100 {
			res.ProgressPercentage = progressPercentage
		}
	}

	res.Status = status
	return s.repo.UpdateResource(ctx, res)
}

func (s *TrackersService) ListResourcesByStudent(ctx context.Context, studentID string) ([]*domain.Resource, error) {
	return s.repo.ListResourcesByStudent(ctx, studentID)
}

func (s *TrackersService) DeleteResource(ctx context.Context, id string) error {
	return s.repo.DeleteResource(ctx, id)
}

// --- School Timetable ---

func (s *TrackersService) SaveTimetable(ctx context.Context, studentID string, entries []*domain.SchoolTimetableEntry) error {
	// Verify student
	_, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Weekday < 1 || entry.Weekday > 7 {
			return errors.New("gün 1 ile 7 arasında olmalı")
		}
		if entry.StartTime == "" || entry.EndTime == "" || entry.SubjectName == "" {
			return errors.New("eksik ders saati veya adı bilgisi")
		}
	}

	return s.repo.SaveTimetable(ctx, entries, studentID)
}

func (s *TrackersService) GetTimetableByStudent(ctx context.Context, studentID string) ([]*domain.SchoolTimetableEntry, error) {
	return s.repo.GetTimetableByStudent(ctx, studentID)
}

// --- Weekly Coaching Plan ---

func (s *TrackersService) CreateWeeklyPlanItem(
	ctx context.Context,
	studentID, coachID string,
	date time.Time,
	start, end, title string,
	subjectID, topicID *string,
	note, url string,
) (*domain.WeeklyPlanItem, error) {
	var resolvedTopicID *string
	if subjectID != nil && *subjectID != "" {
		var err error
		resolvedTopicID, err = s.resolveTopicID(ctx, *subjectID, topicID)
		if err != nil {
			return nil, err
		}
	}

	item := &domain.WeeklyPlanItem{
		StudentID: studentID,
		CoachID:   coachID,
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Title:     title,
		SubjectID: subjectID,
		TopicID:   resolvedTopicID,
		Note:      note,
		URL:       url,
	}

	err := s.repo.CreateWeeklyPlanItem(ctx, item)
	return item, err
}

func (s *TrackersService) UpdateWeeklyPlanItem(
	ctx context.Context,
	id string,
	date time.Time,
	start, end, title string,
	subjectID, topicID *string,
	note, url string,
) error {
	item, err := s.repo.GetWeeklyPlanItemByID(ctx, id)
	if err != nil {
		return err
	}

	var resolvedTopicID *string
	if subjectID != nil && *subjectID != "" {
		var err error
		resolvedTopicID, err = s.resolveTopicID(ctx, *subjectID, topicID)
		if err != nil {
			return err
		}
	}

	item.Date = date
	item.StartTime = start
	item.EndTime = end
	item.Title = title
	item.SubjectID = subjectID
	item.TopicID = resolvedTopicID
	item.Note = note
	item.URL = url

	return s.repo.UpdateWeeklyPlanItem(ctx, item)
}

func (s *TrackersService) DeleteWeeklyPlanItem(ctx context.Context, id string) error {
	return s.repo.DeleteWeeklyPlanItem(ctx, id)
}

func (s *TrackersService) ListWeeklyPlanItems(ctx context.Context, studentID string, start, end time.Time) ([]*domain.WeeklyPlanItem, error) {
	return s.repo.ListWeeklyPlanItems(ctx, studentID, start, end)
}

func (s *TrackersService) CopyWeekPlan(ctx context.Context, studentID string, fromStart, toStart time.Time) error {
	return s.repo.CopyWeek(ctx, studentID, fromStart, toStart)
}

// --- Missing Topics ---

func (s *TrackersService) CreateMissingTopic(
	ctx context.Context,
	studentID, subjectID, topicID, description, priority, status string,
	targetDate *time.Time,
	solutionText, url string,
) (*domain.MissingTopic, error) {
	resolvedTopicID, err := s.resolveTopicID(ctx, subjectID, &topicID)
	if err != nil {
		return nil, err
	}
	if resolvedTopicID == nil {
		return nil, errors.New("konu zorunludur")
	}

	mt := &domain.MissingTopic{
		StudentID:    studentID,
		SubjectID:    subjectID,
		TopicID:      *resolvedTopicID,
		Description:  description,
		Priority:     priority,
		Status:       status,
		TargetDate:   targetDate,
		SolutionText: solutionText,
		URL:          url,
	}

	err = s.repo.CreateMissingTopic(ctx, mt)
	return mt, err
}

func (s *TrackersService) UpdateMissingTopic(
	ctx context.Context,
	id string,
	subjectID, topicID, description, priority, status string,
	targetDate *time.Time,
	solutionText, url string,
) error {
	mt, err := s.repo.GetMissingTopicByID(ctx, id)
	if err != nil {
		return err
	}

	resolvedTopicID, err := s.resolveTopicID(ctx, subjectID, &topicID)
	if err != nil {
		return err
	}
	if resolvedTopicID == nil {
		return errors.New("konu zorunludur")
	}

	mt.SubjectID = subjectID
	mt.TopicID = *resolvedTopicID
	mt.Description = description
	mt.Priority = priority
	mt.Status = status
	mt.TargetDate = targetDate
	mt.SolutionText = solutionText
	mt.URL = url

	return s.repo.UpdateMissingTopic(ctx, mt)
}

func (s *TrackersService) DeleteMissingTopic(ctx context.Context, id string) error {
	return s.repo.DeleteMissingTopic(ctx, id)
}

func (s *TrackersService) ListMissingTopicsByStudent(ctx context.Context, studentID string) ([]*domain.MissingTopic, error) {
	return s.repo.ListMissingTopicsByStudent(ctx, studentID)
}

// --- Goals & Meetings CRUD ---

func (s *TrackersService) CreateGoal(
	ctx context.Context,
	studentID, goalType, title, description string,
	targetVal float64,
	unit string,
	start, target time.Time,
) (*domain.Goal, error) {
	g := &domain.Goal{
		StudentID:    studentID,
		Type:         goalType,
		Title:        title,
		Description:  description,
		TargetValue:  targetVal,
		CurrentValue: 0.0,
		Unit:         unit,
		StartDate:    start,
		TargetDate:   target,
		Status:       "active",
	}

	// Auto calculate initial progress if database contains records
	s.AutoCalculateGoalProgress(ctx, g)

	err := s.repo.CreateGoal(ctx, g)
	return g, err
}

func (s *TrackersService) GetGoalByID(ctx context.Context, id string) (*domain.Goal, error) {
	g, err := s.repo.GetGoalByID(ctx, id)
	if err == nil {
		s.AutoCalculateGoalProgress(ctx, g)
	}
	return g, err
}

func (s *TrackersService) UpdateGoalManualProgress(ctx context.Context, id string, currentVal float64, status string) error {
	g, err := s.repo.GetGoalByID(ctx, id)
	if err != nil {
		return err
	}

	if g.Type != "custom" {
		return errors.New("yalnızca özel hedefler manuel olarak güncellenebilir")
	}

	g.CurrentValue = currentVal
	g.Status = status
	return s.repo.UpdateGoal(ctx, g)
}

func (s *TrackersService) ListGoalsByStudent(ctx context.Context, studentID string) ([]*domain.Goal, error) {
	list, err := s.repo.ListGoalsByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	for _, g := range list {
		s.AutoCalculateGoalProgress(ctx, g)
	}
	return list, nil
}

func (s *TrackersService) DeleteGoal(ctx context.Context, id string) error {
	return s.repo.DeleteGoal(ctx, id)
}

func (s *TrackersService) AutoCalculateGoalProgress(ctx context.Context, g *domain.Goal) {
	if g.Type == "custom" {
		return
	}

	switch g.Type {
	case "question_count":
		// Sum correct answers solved in daily entries during dates
		entries, err := s.repo.ListQuestionSolvingByStudent(ctx, g.StudentID, &g.StartDate, &g.TargetDate)
		if err == nil {
			sum := 0
			for _, e := range entries {
				sum += (e.Correct + e.Incorrect + e.Blank)
			}
			g.CurrentValue = float64(sum)
		}

	case "exam_net":
		// Maximum net score achieved in trial exams in this range
		exams, err := s.examRepo.ListExamsByStudent(ctx, g.StudentID, "", &g.StartDate, &g.TargetDate)
		if err == nil {
			maxNet := 0.0
			for _, e := range exams {
				if e.TotalNet > maxNet {
					maxNet = e.TotalNet
				}
			}
			g.CurrentValue = maxNet
		}

	case "resource_completion":
		// Find completion percentage of linked books/resources
		resources, err := s.repo.ListResourcesByStudent(ctx, g.StudentID)
		if err == nil {
			totalProgress := 0.0
			count := 0.0
			for _, r := range resources {
				totalProgress += float64(r.ProgressPercentage)
				count++
			}
			if count > 0 {
				g.CurrentValue = totalProgress / count
			}
		}
	}

	// Auto check achieved status
	if g.CurrentValue >= g.TargetValue {
		g.Status = "achieved"
	} else if time.Now().After(g.TargetDate) && g.Status == "active" {
		g.Status = "failed"
	}
}

// --- Coaching Meetings ---

func (s *TrackersService) CreateMeeting(
	ctx context.Context,
	studentID, coachID string,
	date time.Time,
	duration int,
	notes, url string,
	nextMeeting *time.Time,
) (*domain.Meeting, error) {
	m := &domain.Meeting{
		StudentID:       studentID,
		CoachID:         coachID,
		MeetingDate:     date,
		DurationMinutes: duration,
		Notes:           notes,
		MeetingURL:      url,
		NextMeetingDate: nextMeeting,
		Status:          "planned",
	}

	err := s.repo.CreateMeeting(ctx, m)
	return m, err
}

func (s *TrackersService) UpdateMeeting(
	ctx context.Context,
	id string,
	date time.Time,
	duration int,
	notes, url string,
	nextMeeting *time.Time,
	status string,
) error {
	m, err := s.repo.GetMeetingByID(ctx, id)
	if err != nil {
		return err
	}

	m.MeetingDate = date
	m.DurationMinutes = duration
	m.Notes = notes
	m.MeetingURL = url
	m.NextMeetingDate = nextMeeting
	m.Status = status

	return s.repo.UpdateMeeting(ctx, m)
}

func (s *TrackersService) DeleteMeeting(ctx context.Context, id string) error {
	return s.repo.DeleteMeeting(ctx, id)
}

func (s *TrackersService) ListMeetingsByStudent(ctx context.Context, studentID string) ([]*domain.Meeting, error) {
	return s.repo.ListMeetingsByStudent(ctx, studentID)
}

func (s *TrackersService) ListMeetingsByCoach(ctx context.Context, coachID string) ([]*domain.Meeting, error) {
	return s.repo.ListMeetingsByCoach(ctx, coachID)
}
