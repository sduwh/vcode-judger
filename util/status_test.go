package util

import (
	"github.com/sduwh/vcode-judger/consts"
	"testing"
)

type statusTestCase struct {
	CaseValue  string
	CaseAnswer string
}

func TestStatusMapper(t *testing.T) {
	caseList := []statusTestCase{
		{
			"judge",
			consts.StatusJudging,
		},
		{
			"accept",
			consts.StatusAccepted,
		},
		{
			"timeLimit",
			consts.StatusTimeLimitExceeded,
		},
		{
			"memoryLimit",
			consts.StatusMemoryLimitExceeded,
		},
		{
			"Compile error",
			consts.StatusCompileError,
		},
	}
	for _, nowCase := range caseList {
		result := JudgeStatusMapper(nowCase.CaseValue)
		if result != nowCase.CaseAnswer {
			t.Errorf("Identify the state error, case: %+v, result: %+v", nowCase.CaseValue, result)
		} else {
			t.Logf("Identify the state success, case: %+v, result: %+v", nowCase.CaseValue, result)
		}
	}
}
