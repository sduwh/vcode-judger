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
	ID         string `json:"id"`
	ProviderOJ string `json:"provider_oj"`
	ProviderID string `json:"provider_id"`
	Language   string `json:"language"`
	Code       string `json:"code"`
}

type JudgeStatus struct {
	TaskID       string `json:"task_id"`
	SubmitID     string `json:"submit_id"`
	Status       string `json:"status"`
	TimeUsed     int64  `json:"time_used"`
	MemoryUsed   int64  `json:"time_used"`
	CompileError string `json:"compile_error"`
}
