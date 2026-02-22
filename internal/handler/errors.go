package handler

import "errors"

var (
	errAIDisabled  = errors.New("AI features are disabled. Use --ai=true to enable")
	errNoLogGroup  = errors.New("no CloudWatch log group found in task definition")
	errNoLogsFound = errors.New("no recent logs found")
)
