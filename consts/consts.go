package consts

import "time"

const (
	Version           = "0.1.0"
	MaxStatusWaitTime = 5 * time.Minute
)

// Topics
const (
	TopicTask       = "vcode-judge-task"
	TopicRemoteTask = "vcode-judge-remote-task"
	TopicStatus     = "vcode-judge-status"
)

// Providers
const (
	RemotePOJ  = "POJ"
	RemoteHDU  = "HDU"
	RemoteSDUT = "SDUT"
)

// Languages
const (
	LanguageC    = "C"
	LanguageCPP  = "CPP"
	LanguageJava = "JAVA"
)

// Statuses
const (
	StatusJudging             = "Judging"
	StatusAccepted            = "Accepted"
	StatusPresentationError   = "Presentation Error"
	StatusTimeLimitExceeded   = "Time Limit Exceeded"
	StatusMemoryLimitExceeded = "Memory Limit Exceeded"
	StatusWrongAnswer         = "Wrong Answer"
	StatusRuntimeError        = "Runtime Error"
	StatusOutputLimitExceeded = "Output Limit Exceeded"
	StatusCompileError        = "Compile Error"
)
