package logs_converter

const (
	fldLogzioTimestamp = "@timestamp"
	fldLambdaTime = "time"
	fldLogzioType = "type"
	fldLambdaType = "type"
	fldLogzioLambdaType = "lambda.log.type"
	fldLambdaRecord = "record"
	fldLogzioMsg = "message"
	fldLogzioLambdaRecord = "lambda.record"

	extensionType = "lambda-extension-logs"
)

// ConvertLambdaLogToLogzioLog converts a log that was sent from AWS Logs API to a log in a Logz.io format
func ConvertLambdaLogToLogzioLog(lambdaLog map[string]interface{}) map[string]interface{}{
	logzioLog := make(map[string]interface{})
	logzioLog[fldLogzioTimestamp] = lambdaLog[fldLambdaTime]
	logzioLog[fldLogzioType] = extensionType
	logzioLog[fldLogzioLambdaType] = lambdaLog[fldLambdaType]

	switch lambdaLog[fldLambdaRecord].(type) {
	case string:
		logzioLog[fldLogzioMsg] = lambdaLog[fldLambdaRecord]
	default:
		logzioLog[fldLogzioLambdaRecord] = lambdaLog[fldLambdaRecord]
	}

	return logzioLog
}
