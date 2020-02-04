package main

type MessageTask struct {
	ID          string `json:"id"`
	CaseID      string `json:"case_id"`
	TimeLimit   uint64 `json:"time_limit"`
	MemoryLimit uint64 `json:"memory_limit"`
	Language    string `json:"language"`
	Code        string `json:"code"`
}

type MessageRemoteTask struct {
	ID         string `json:"id"`
	ProviderOJ string `json:"provider_oj"`
	ProviderID string `json:"provider_id"`
	Language   string `json:"language"`
	Code       string `json:"code"`
}

type MessageStatus struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}
