package utils_test

import (
	"github.com/stretchr/testify/assert"
	"logzio-lambda-extensions-logs/utils"
	"os"
	"testing"
)

func clearTestEnv() {
	os.Unsetenv("GROK_PATTERNS")
	os.Unsetenv("LOGS_FORMAT")
	os.Unsetenv("CUSTOM_FIELDS")
	os.Unsetenv("JSON_FIELDS_UNDER_ROOT")
	os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")
	os.Unsetenv("AWS_REGION")
}

func TestConverterSimpleLog(t *testing.T) {
	clearTestEnv()
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "this is a simple log\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Equal(t, lambdaLog[utils.FldLambdaRecord], logzioLog[utils.FldLogzioMsg])
}

func TestConverterSimpleJsonLog(t *testing.T) {
	clearTestEnv()
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "{\"foo\": \"bar\"}\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Equal(t, "bar", logzioLog[utils.FldLogzioMsgNested].(map[string]interface{})["foo"])
}

func TestConverterSimpleJsonLogAndJsonFieldsUnderRoot(t *testing.T) {
	clearTestEnv()
	os.Setenv("JSON_FIELDS_UNDER_ROOT", "true")
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "{\"message\": \"hello\", \"some\": {\"metadata\": \"object\"}}\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Equal(t, "hello", logzioLog["message"])
	assert.Equal(t, "object", logzioLog["some"].(map[string]interface{})["metadata"])
}

func TestConverterGrokFormattedLog(t *testing.T) {
	clearTestEnv()
	os.Setenv("GROK_PATTERNS", "{\"app_name\":\"cool app\",\"my_message\":\".*\"}")
	os.Setenv("LOGS_FORMAT", "%{app_name:my_app} : %{my_message:my_message}")
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "cool app : this is a formatted log\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Equal(t, "cool app", logzioLog["my_app"])
	assert.Equal(t, "this is a formatted log", logzioLog["my_message"])
}

func TestConverterGrokFormattedLogWithJson(t *testing.T) {
	clearTestEnv()
	os.Setenv("GROK_PATTERNS", "{\"app_name\":\"cool app\",\"my_message\":\".*\"}")
	os.Setenv("LOGS_FORMAT", "%{app_name:my_app} : %{my_message:my_message}")
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "cool app : {\"foo\": \"bar\"}\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Equal(t, "cool app", logzioLog["my_app"])
	assert.Equal(t, "bar", logzioLog["my_message"].(map[string]interface{})["foo"])
}

func TestConverterGrokFormattedLogIncorrectLogsFormat(t *testing.T) {
	clearTestEnv()
	os.Setenv("GROK_PATTERNS", "{\"app_name\":\"cool app\",\"my_message\":\".*\"}")
	os.Setenv("LOGS_FORMAT", "%{app_name:my_app} = %{my_message:my_message}")
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "cool app : this is a formatted log\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Nil(t, logzioLog["my_app"])
	assert.Nil(t, logzioLog["my_message"])
	assert.NotNil(t, logzioLog[utils.FldLogzioMsg])
	assert.Equal(t, "cool app : this is a formatted log\n", logzioLog[utils.FldLogzioMsg])
}

func TestConverterGrokFormattedLogIncorrectGrokPattern(t *testing.T) {
	clearTestEnv()
	os.Setenv("GROK_PATTERNS", "{\"app_name\":\"some app\",\"my_message\":\".*\"}")
	os.Setenv("LOGS_FORMAT", "%{app_name:my_app} : %{my_message:my_message}")
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "cool app : this is a formatted log\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Nil(t, logzioLog["my_app"])
	assert.Nil(t, logzioLog["my_message"])
	assert.NotNil(t, logzioLog[utils.FldLogzioMsg])
	assert.Equal(t, "cool app : this is a formatted log\n", logzioLog[utils.FldLogzioMsg])
}

func TestConverterGrokFormattedLogNoGrokPattern(t *testing.T) {
	clearTestEnv()
	os.Setenv("LOGS_FORMAT", "%{app_name:my_app} : %{my_message:my_message}")
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "cool app : this is a formatted log\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Nil(t, logzioLog["my_app"])
	assert.Nil(t, logzioLog["my_message"])
	assert.NotNil(t, logzioLog[utils.FldLogzioMsg])
	assert.Equal(t, "cool app : this is a formatted log\n", logzioLog[utils.FldLogzioMsg])
}

func TestConverterGrokFormattedLogNoLogsFormat(t *testing.T) {
	clearTestEnv()
	os.Setenv("GROK_PATTERNS", "{\"app_name\":\"cool app\",\"my_message\":\".*\"}")
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "cool app : this is a formatted log\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaLog[utils.FldLambdaTime], logzioLog[utils.FldLogzioTimestamp])
	assert.Equal(t, lambdaLog[utils.FldLambdaType], logzioLog[utils.FldLogzioLambdaType])
	assert.Nil(t, logzioLog["my_app"])
	assert.Nil(t, logzioLog["my_message"])
	assert.NotNil(t, logzioLog[utils.FldLogzioMsg])
	assert.Equal(t, "cool app : this is a formatted log\n", logzioLog[utils.FldLogzioMsg])
}

func TestConverterAddAwsMetadata(t *testing.T) {
	clearTestEnv()
	lambdaName := "my lambda"
	region := "us-east-1"
	os.Setenv("AWS_LAMBDA_FUNCTION_NAME", lambdaName)
	os.Setenv("AWS_REGION", region)
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "{\"foo\": \"bar\"}\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, lambdaName, logzioLog[utils.FldLogzioLambdaName])
	assert.Equal(t, region, logzioLog[utils.FldLogzioAwsRegion])
}

func TestConverterCustomFields(t *testing.T) {
	clearTestEnv()
	custom := "hello=world,hola=mundo"
	os.Setenv("CUSTOM_FIELDS", custom)
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "{\"foo\": \"bar\"}\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.Equal(t, "world", logzioLog["hello"])
	assert.Equal(t, "mundo", logzioLog["hola"])
}

func TestConverterCustomFieldsKeyExistInLog(t *testing.T) {
	clearTestEnv()
	custom := "my_message=world,hola=mundo"
	os.Setenv("CUSTOM_FIELDS", custom)
	utils.InitConfig()
	lambdaLog := map[string]interface{}{
		utils.FldLambdaTime:   "2021-11-11T08:28:16.870Z",
		utils.FldLambdaType:   "function",
		utils.FldLambdaRecord: "{\"foo\": \"bar\"}\n",
	}

	logzioLog := utils.ConvertLambdaLogToLogzioLog(lambdaLog)
	assert.NotNil(t, logzioLog)
	assert.NotZero(t, len(logzioLog))
	assert.NotEqual(t, "world", logzioLog["message"])
	assert.Equal(t, "mundo", logzioLog["hola"])
}
