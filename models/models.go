package models

type JudgeTask struct {
	ID          string `json:"id"`
	CaseID      string `json:"case_id"`
	TimeLimit   uint64 `json:"time_limit"`
	MemoryLimit uint64 `json:"memory_limit"`
	Language    string `json:"language"`
	Code        string `json:"code"`
}

type RemoteJudgeTask struct {
	ID         string `json:"submit_id"`
	ProviderOJ string `json:"origin"`
	ProviderID string `json:"key"`
	Language   string `json:"language"`
	Code       string `json:"code"`
}

type JudgeStatus struct {
	TaskID       string `json:"task_id"`   //vcode's submit
	SubmitID     string `json:"submit_id"` // origin oj's submit
	Status       string `json:"status"`
	TimeUsed     int64  `json:"time_used"`
	MemoryUsed   int64  `json:"memory_used"`
	CompileError string `json:"compile_error"`
}
