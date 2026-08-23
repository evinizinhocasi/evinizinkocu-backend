package domain

type MonthlyReportSummary struct {
	TotalQuestionsSolved   int     `json:"totalQuestionsSolved"`
	TotalCorrect           int     `json:"totalCorrect"`
	TotalWrong             int     `json:"totalWrong"`
	TotalEmpty             int     `json:"totalEmpty"`
	TotalNet               float64 `json:"totalNet"`
	TotalHomeworkAssigned  int     `json:"totalHomeworkAssigned"`
	TotalHomeworkCompleted int     `json:"totalHomeworkCompleted"`
	HomeworkCompletionRate float64 `json:"homeworkCompletionRate"`
	WrongQuestionsLogged   int     `json:"wrongQuestionsLogged"`
	WrongQuestionsResolved int     `json:"wrongQuestionsResolved"`
	TrialExamsCount        int     `json:"trialExamsCount"`
	AverageTrialNet        float64 `json:"averageTrialNet"`
	MeetingsCount          int     `json:"meetingsCount"`
}

type SubjectReportDetail struct {
	SubjectName     string  `json:"subjectName"`
	QuestionsSolved int     `json:"questionsSolved"`
	Correct         int     `json:"correct"`
	Wrong           int     `json:"wrong"`
	Empty           int     `json:"empty"`
	Net             float64 `json:"net"`
	TasksAssigned   int     `json:"tasksAssigned"`
	TasksCompleted  int     `json:"tasksCompleted"`
}

type TrialExamReportDetail struct {
	ExamTitle string  `json:"examTitle"`
	ExamDate  string  `json:"examDate"`
	TotalNet  float64 `json:"totalNet"`
	Score     float64 `json:"score"`
	Rank      int     `json:"rank"`
}

type MonthlyReportData struct {
	StudentID         string                  `json:"studentId"`
	StudentName       string                  `json:"studentName"`
	ClassLevel        string                  `json:"classLevel"`
	StudyTrack        string                  `json:"studyTrack"`
	ExamTypeName      string                  `json:"examTypeName"`
	CoachName         string                  `json:"coachName"`
	TargetSchoolName  string                  `json:"targetSchoolName"`
	TargetSchoolPhoto string                  `json:"targetSchoolPhoto"`
	Year              int                     `json:"year"`
	Month             int                     `json:"month"`
	MonthName         string                  `json:"monthName"`
	Summary           MonthlyReportSummary    `json:"summary"`
	SubjectBreakdown  []SubjectReportDetail   `json:"subjectBreakdown"`
	TrialExams        []TrialExamReportDetail `json:"trialExams"`
}
