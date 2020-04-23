package util

import (
	"github.com/sduwh/vcode-judger/consts"
	"regexp"
)

const (
	StatusJudgingRegular             = "[Jj]udg.*|.*ing"
	StatusAcceptedRegular            = "[Aa]ccept.*|[Ss]uccess"
	StatusPresentationErrorRegular   = "[Pp]resentation ?[Ee]rror"
	StatusTimeLimitExceededRegular   = "[Tt]ime ?[Ll]imit"
	StatusMemoryLimitExceededRegular = "[Mm]emory ?[Ll]imit"
	StatusWrongAnswerRegular         = "[Ww]rong ?[Aa]nswer"
	StatusRuntimeErrorRegular        = "[Rr]untime ?[Ee]rror"
	StatusOutputLimitExceededRegular = "[Oo]utput ?[Ll]imit"
	StatusCompileErrorRegular        = "[Cc]ompile ?[Ee]rror"
)

func JudgeStatusMapper(status string) string {
	result := false
	if result, _ = regexp.MatchString(StatusJudgingRegular, status); result {
		return consts.StatusJudging
	} else if result, _ = regexp.MatchString(StatusAcceptedRegular, status); result {
		return consts.StatusAccepted
	} else if result, _ = regexp.MatchString(StatusPresentationErrorRegular, status); result {
		return consts.StatusPresentationError
	} else if result, _ = regexp.MatchString(StatusTimeLimitExceededRegular, status); result {
		return consts.StatusTimeLimitExceeded
	} else if result, _ = regexp.MatchString(StatusMemoryLimitExceededRegular, status); result {
		return consts.StatusMemoryLimitExceeded
	} else if result, _ = regexp.MatchString(StatusRuntimeErrorRegular, status); result {
		return consts.StatusRuntimeError
	} else if result, _ = regexp.MatchString(StatusWrongAnswerRegular, status); result {
		return consts.StatusWrongAnswer
	} else if result, _ = regexp.MatchString(StatusOutputLimitExceededRegular, status); result {
		return consts.StatusOutputLimitExceeded
	} else if result, _ = regexp.MatchString(StatusCompileErrorRegular, status); result {
		return consts.StatusCompileError
	} else {
		return consts.StatusUnknownError
	}
}
