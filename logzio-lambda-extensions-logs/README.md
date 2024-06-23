# Logzio-lambda-extensions-logs

Lambda extensions enable tools to integrate deeply into the Lambda execution environment to control and participate in Lambda’s lifecycle.
To read more about Lambda Extensions, [click here](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html).  
The Logz.io Lambda extension for logs uses the AWS Extensions API and [AWS Logs API](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-logs-api.html), and sends your Lambda Function Logs directly to your Logz.io account.

This repo is based on the [AWS lambda extensions sample](https://github.com/aws-samples/aws-lambda-extensions).

This extension is written in Go, but can be run with runtimes that support [extensions](https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html).

### Prerequisites

* Lambda function with [supported runtime](https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html) for extensions.
* AWS Lambda limitations: A function can use up to five layers at a time. The total unzipped size of the function and all layers cannot exceed the unzipped deployment package size limit of 250 MB.

### Important notes

* If the extension doesn't have enough time to receive logs from AWS Logs API, it may send the logs in the next invocation of the Lambda function.
  So if you want that all the logs are sent by the end of your function's run, you'll need to add at the end of your Lambda function code a sleep interval that will allow the extension enough time to do its job.
* Due to [Lambda's execution environment lifecycle](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-context.html), the extension is being invoked on two events - `INVOKE` and `SHUTDOWN`.
  This means that if your Lambda function goes into the `SHUTDOWN` phase, the extension will run and if there are logs in its queue, it will send them.

### Extension deployment options

You can deploy the extension via:
* [AWS CLI](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#deploying-logzio-logs-extension-via-the-aws-cli).
* [AWS Management Console](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#deploying-logzio-log-extensions-via-the-aws-management-console).

### Deploying Logz.io logs extension via the AWS CLI

##### Deploy the extension and configuration

If you haven't done it already, [install](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) and [configure](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) the AWS CLI.

Add the layer to your function and configure the environment variables using the following command:

```shell
aws lambda update-function-configuration \
    --function-name <<FUNCTION-NAME>> \
    --layers <<LAYERS>> \
    --environment "Variables={<<ENV-VARS>>}"
```

**Note:** this command overwrites the existing function configuration. If you already have your own layers and environment variables for your function, list them as well.

| Placeholder | Description                                                                                                                                                                                                                                                                                                                                                                                                             |
|---|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `<<FUNCTION-NAME>>` | Name of the Lambda function you want to monitor.                                                                                                                                                                                                                                                                                                                                                                        |
| `<<LAYERS>>` | A space-separated list of function layers to add to the function's execution environment. Specify each layer by its ARN, including the version.  For the ARN, see the [**ARNs** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#arns).                                                                                                                                |
| `<<ENV-VARS>>`  | Key-value pairs containing environment variables that are accessible from function code during execution. Should appear in the following format: `KeyName1=string,KeyName2=string`.  For a list of all the environment variables for the extension, see the [**Lambda environment variables** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#environment-variables). |

##### Run the function

Use the following command. It may take more than one run of the function for the logs to start shipping to your Logz.io account.

```shell
aws lambda update-function-configuration \
    --function-name <<FUNCTION-NAME>> \
    --layers [] \
    --environment "Variables={}"
```

Your lambda logs will appear under the type `lambda-extension-logs`.

**NOTE:** This command overwrites the existing function configuration. If you already have your own layers and environment variables for your function, include them in the list.


#### Deleting the extension

To delete the extension and its environment variables, use the following command:

```shell
aws lambda update-function-configuration \
    --function-name some-func \
    --layers [] \
    --environment "Variables={}"
```

**NOTE:** This command overwrites the existing function configuration. If you already have your own layers and environment variables for your function, include them in the list.

### Deploying Logz.io log extensions via the AWS Management Console

##### Add the extension to your Lambda Function

1. In the Lambda Functions screen, choose the function you want to monitor.
   ![Pick lambda function](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lambda_extensions/lambda-x_1-1.jpg)

2. In the page for the function, scroll down to the `Layers` section and choose `Add Layer`.
   ![Add layer](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lambda_extensions/lambda-x_1-2.jpg)

3. Select the `Specify an ARN` option, then choose the ARN of the extension with the region code that matches your Lambda Function region from the [**ARNs** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#arns), and click the `Add` button.
   ![Add ARN extension](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lambda_extensions/lambda-x_1-3.jpg)

##### Configure the extension parameters

Add the environment variables to the function, according to the [**Environment variables** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#environment-variables).

##### Run the function

Run the function. It may take more than one run of the function for the logs to start shipping to your Logz.io account.

Your lambda logs will appear under the type `lambda-extension-logs`.

#### Deleting the extension

- To delete the **extension layer**: In your function page, go to the **layers** panel. Click `edit`, select the extension layer, and click `save`.
- To delete the extension's **environment variables**: In your function page, select the `Configuration` tab, select `Environment variables`, click `edit`, and remove the variables that you added for the extension.

### Parsing logs

By default, the extension sends the logs as strings.  
If your logs are formatted, and you wish to parse them to separate fields, the extension will use the [grok library](https://github.com/vjeantet/grok) to parse grok patterns.
You can see all the pre-built grok patterns (for example `COMMONAPACHELOG` is already a known pattern in the library) [here](https://github.com/vjeantet/grok/tree/master/patterns).
If you need to use a custom pattern, you can use the environment variables `GROK_PATTERNS` and `LOGS_FORMAT`.

#### Example

For logs that are formatted like this:

```python
<<timestamp>> <<app_name>>: <<message>>
# Examples
May 04 2024 10:48:34.244 my_app: an awesome message
May 04 2024 10:50:46.532 logzio_sender: Successfully sent bulk to logz.io, size: 472
```

In Logz.io we wish to have `timestamp`, `app_name` and `message` in their own fields.  
To do so, we'll set the environment variables as follows:

##### GROK_PATTERNS

The `GROK_PATTERNS` variable contains definitions of custom grok patterns and should be in a JSON format.   
- key - is the custom pattern name. 
- value - the regex that captures the pattern.

In our example:
- `timestamp` - matching the regex `\w+ \d{2} \d{4} \d{2}:\d{2}:\d{2}\.\d{3}`.
- `app_name` - always a not space, so matching `\S+`.
- `message` -  have strings containing whitespaces, letters and numbers. So matching `.*`.

For the regex that matches `app_name` and `message` there are built in grok patterns (we'll see in `LOGS_FORMAT` explanation), so we only need to define custom pattern for our `timestamp`.  
Meaning we can set `GROK_PATTERNS` as: 
``` json
{"MY_CUSTOM_TIMESTAMP":"\\w+ \\d{2} \\d{4} \\d{2}:\\d{2}:\\d{2}\\.\\d{3}"}
```

##### LOGS_FORMAT

The `LOGS_FORMAT` variable contains the full grok patternt that will match the format of the logs, using known patterns and the custom patterns that were defined in `GROK_PATTERNS` (if defined).  
The variable should be in a grok format: 
```
%{GROK_PATTERN_NAME:WANTED_FIELD_NAME}
```
**Note**: the `WANTED_FIELD_NAME` cannot contain a dot (`.`) in it.

In our example: 
- `timestamp` - matching the custom pattern we defined previously `MY_CUSTOM_TIMESTAMP`.
- `app_name` - is matching the known grok pattern `NOTSPACE`.
- `message` -  is matching the known grok pattern `GREEDYDATA`.

So we will set `LOGS_FORMAT` as: 
```
^%{MY_CUSTOM_TIMESTAMP:timestamp} %{NOTSPACE:app_name}: %{GREEDYDATA:message}
```

The log example from above: 
```
May 04 2024 10:48:34.244 my_app: an awesome message
```
Will be parsed to look like this:

```
timestamp: May 04 2024 10:48:34.244
app_name: my_app
message: an awesome message
```

This project uses an external module for its Grok parsing. To learn more about it, see the [grok library repo](https://github.com/vjeantet/grok).

### Nested fields

As of v0.2.0, by default, the extension can detect if a log is in a JSON format, and to parse the fields to appear as nested fields in the Logz.io app.
For example, the following log:

```
{ "foo": "bar", "field2": "val2" }
```

Will appear under the fields:
```
message_nested.foo: bar
message_nested.field2: val2
```

As of v0.3.3, to have the fields nested under the root (instead of under `message_nested`), set the `JSON_FIELDS_UNDER_ROOT` environment variable as `true`.  
It is useful in cases where the passed object is in fact meant to be that of a message plus metadata fields.  
For example, the following log:

```
{ "message": "hello", "foo": "bar" }
```

Will appear under the fields:
```
message: hello
foo: bar
```

**Note:** The user must insert a valid JSON. Sending a dictionary or any key-value data structure that is not in a JSON format will cause the log to be sent as a string.

### Environment Variables

| Name                     | Description                                                                                                                                                                                                                                                                                                                                 | Required/Default |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------|
| `LOGZIO_LOGS_TOKEN`      | Your Logz.io log shipping [token](https://app.logz.io/#/dashboard/settings/manage-tokens/data-shipping).                                                                                                                                                                                                                                    | Required         |
| `LOGZIO_LISTENER`        | Your  Logz.io listener address, with port 8070 (http) or 8071 (https). For example: `https://listener.logz.io:8071`                                                                                                                                                                                                                         | Required         |
| `LOGS_EXT_LOG_LEVEL`     | Log level of the extension. Can be set to one of the following: `debug`, `info`, `warn`, `error`, `fatal`, `panic`.                                                                                                                                                                                                                         | Default: `info`  |
| `ENABLE_PLATFORM_LOGS`   | The platform log captures runtime or execution environment errors. Set to `true` if you wish the platform logs will be shipped to your Logz.io account.                                                                                                                                                                                     | Default: `false` |
| `GROK_PATTERNS`          | Must be set with `LOGS_FORMAT`. Use this if you want to parse your logs into fields. A minified JSON list that contains the field name and the regex that will match the field. To understand more see the [parsing logs](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#parsing-logs) section. | -                |
| `LOGS_FORMAT`            | Must be set with `GROK_PATTERNS`. Use this if you want to parse your logs into fields. The format in which the logs will appear, in accordance to grok conventions. To understand more see the [parsing logs](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#parsing-logs) section.             | -                |
| `CUSTOM_FIELDS`          | Include additional fields with every message sent, formatted as `fieldName1=fieldValue1,fieldName2=fieldValue2` (**NO SPACES**). A custom key that clashes with a key from the log itself will be ignored.                                                                                                                                  | -                |
| `JSON_FIELDS_UNDER_ROOT` | If you log Json messages and would like the fields to be nested at the root of the log, instead of under `message_nested`.                                                                                                                                                                                                                  | Default: `False` |


### ARNs
## AMD64 Architecture

| Region Name               | Region Code      | AWS ARN                                                                         |
|---------------------------|------------------|---------------------------------------------------------------------------------|
| US East (N. Virginia)     | `us-east-1`      | `arn:aws:lambda:us-east-1:486140753397:layer:LogzioLambdaExtensionLogs:14`      |
| US East (Ohio)            | `us-east-2`      | `arn:aws:lambda:us-east-2:486140753397:layer:LogzioLambdaExtensionLogs:14`      |
| US West (N. California)   | `us-west-1`      | `arn:aws:lambda:us-west-1:486140753397:layer:LogzioLambdaExtensionLogs:14`      |
| US West (Oregon)          | `us-west-2`      | `arn:aws:lambda:us-west-2:486140753397:layer:LogzioLambdaExtensionLogs:13`      |
| Europe (Frankfurt)        | `eu-central-1`   | `arn:aws:lambda:eu-central-1:486140753397:layer:LogzioLambdaExtensionLogs:15`   |
| Europe (Ireland)          | `eu-west-1`      | `arn:aws:lambda:eu-west-1:486140753397:layer:LogzioLambdaExtensionLogs:13`      |
| Europe (Stockholm)        | `eu-north-1`     | `arn:aws:lambda:eu-north-1:486140753397:layer:LogzioLambdaExtensionLogs:14`     |
| Asia Pacific (Sydney)     | `ap-southeast-2` | `arn:aws:lambda:ap-southeast-2:486140753397:layer:LogzioLambdaExtensionLogs:14` |
| Canada (Central)          | `ca-central-1`   | `arn:aws:lambda:ca-central-1:486140753397:layer:LogzioLambdaExtensionLogs:15`   |
| South America (São Paulo) | `sa-east-1`      | `arn:aws:lambda:sa-east-1:486140753397:layer:LogzioLambdaExtensionLogs:16`      |
| Asia Pacific (Tokyo)      | `ap-northeast-1` | `arn:aws:lambda:ap-northeast-1:486140753397:layer:LogzioLambdaExtensionLogs:10` |
| Asia Pacific (Singapore)  | `ap-southeast-1` | `arn:aws:lambda:ap-southeast-1:486140753397:layer:LogzioLambdaExtensionLogs:11` |
| Asia Pacific (Mumbai)     | `ap-south-1`     | `arn:aws:lambda:ap-south-1:486140753397:layer:LogzioLambdaExtensionLogs:10`     |
| Asia Pacific (Osaka)      | `ap-northeast-3` | `arn:aws:lambda:ap-northeast-3:486140753397:layer:LogzioLambdaExtensionLogs:11` |
| Asia Pacific (Seoul)      | `ap-northeast-2` | `arn:aws:lambda:ap-northeast-2:486140753397:layer:LogzioLambdaExtensionLogs:11` |
| Europe (London)           | `eu-west-2`      | `arn:aws:lambda:eu-west-2:486140753397:layer:LogzioLambdaExtensionLogs:12`      |
| Europe (Paris)            | `eu-west-3`      | `arn:aws:lambda:eu-west-3:486140753397:layer:LogzioLambdaExtensionLogs:11`      |

## ARM64 Architecture
| Region Name               | Region Code      | AWS ARN                                                                           |
|---------------------------|------------------|-----------------------------------------------------------------------------------|
| US East (N. Virginia)     | `us-east-1`      | `arn:aws:lambda:us-east-1:486140753397:layer:LogzioLambdaExtensionLogsArm:6`      |
| US East (Ohio)            | `us-east-2`      | `arn:aws:lambda:us-east-2:486140753397:layer:LogzioLambdaExtensionLogsArm:6`      |
| US West (N. California)   | `us-west-1`      | `arn:aws:lambda:us-west-1:486140753397:layer:LogzioLambdaExtensionLogsArm:6`      |
| US West (Oregon)          | `us-west-2`      | `arn:aws:lambda:us-west-2:486140753397:layer:LogzioLambdaExtensionLogsArm:5`      |
| Europe (Frankfurt)        | `eu-central-1`   | `arn:aws:lambda:eu-central-1:486140753397:layer:LogzioLambdaExtensionLogsArm:5`   |
| Europe (Ireland)          | `eu-west-1`      | `arn:aws:lambda:eu-west-1:486140753397:layer:LogzioLambdaExtensionLogsArm:6`      |
| Europe (Stockholm)        | `eu-north-1`     | `arn:aws:lambda:eu-north-1:486140753397:layer:LogzioLambdaExtensionLogsArm:6`     |
| Asia Pacific (Sydney)     | `ap-southeast-2` | `arn:aws:lambda:ap-southeast-2:486140753397:layer:LogzioLambdaExtensionLogsArm:5` |
| Canada (Central)          | `ca-central-1`   | `arn:aws:lambda:ca-central-1:486140753397:layer:LogzioLambdaExtensionLogsArm:5`   |
| South America (São Paulo) | `sa-east-1`      | `arn:aws:lambda:sa-east-1:486140753397:layer:LogzioLambdaExtensionLogsArm:6`      |
| Asia Pacific (Tokyo)      | `ap-northeast-1` | `arn:aws:lambda:ap-northeast-1:486140753397:layer:LogzioLambdaExtensionLogsArm:6` |
| Asia Pacific (Singapore)  | `ap-southeast-1` | `arn:aws:lambda:ap-southeast-1:486140753397:layer:LogzioLambdaExtensionLogsArm:6` |
| Asia Pacific (Mumbai)     | `ap-south-1`     | `arn:aws:lambda:ap-south-1:486140753397:layer:LogzioLambdaExtensionLogsArm:5`     |
| Asia Pacific (Osaka)      | `ap-northeast-3` | `arn:aws:lambda:ap-northeast-3:486140753397:layer:LogzioLambdaExtensionLogsArm:6` |
| Asia Pacific (Seoul)      | `ap-northeast-2` | `arn:aws:lambda:ap-northeast-2:486140753397:layer:LogzioLambdaExtensionLogsArm:6` |
| Europe (London)           | `eu-west-2`      | `arn:aws:lambda:eu-west-2:486140753397:layer:LogzioLambdaExtensionLogsArm:5`      |
| Europe (Paris)            | `eu-west-3`      | `arn:aws:lambda:eu-west-3:486140753397:layer:LogzioLambdaExtensionLogsArm:6`      |

### Lambda extension versions

| Version | Supported Runtimes                                                                                                                                                                                                       |
|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.3.4   | `.NET 6`, `.NET 8`, `provided.al2`, `provided.al2023`, `Java 8`, `Java 11`, `Java 17`, `Node.js 16`, `Node.js 18`, `Python 3.8`, `Python 3.9`, `Python 3.10`, `Python 3.11`, `Python 3.12`, `Ruby 3.2`, `Custom Runtime` |
| 0.3.3   | `.NET 6`, `.NET 8`, `provided.al2`, `provided.al2023`, `Java 8`, `Java 11`, `Java 17`, `Node.js 16`, `Node.js 18`, `Python 3.8`, `Python 3.9`, `Python 3.10`, `Python 3.11`, `Python 3.12`, `Ruby 3.2`, `Custom Runtime` |
| 0.3.2   | `.NET 6`, `Go 1.x`, `Java 17`, `Node.js 18`, `Python 3.11`, `Ruby 3.2`, `Java 11`, `Java 8`, `Node.js 16`, `Python 3.10`, `Python 3.9`, `Python 3.8`, `Ruby 2.7`, `Custom Runtime`                                       |
| 0.3.1   | All runtimes                                                                                                                                                                                                             |
| 0.3.0   | `.NET Core 3.1`, `Java 11`, `Java 8`, `Node.js 14.x`, `Node.js 12.x`, `Python 3.9`, `Python 3.8`, `Python 3.7`, `Ruby 2.7`, `Custom runtime`                                                                             |
| 0.2.0   | `.NET Core 3.1`, `Java 11`, `Java 8`, `Node.js 14.x`, `Node.js 12.x`, `Python 3.9`, `Python 3.8`, `Python 3.7`, `Ruby 2.7`, `Custom runtime`                                                                             |
| 0.1.0   | `.NET Core 3.1`, `Java 11`, `Java 8`, `Node.js 14.x`, `Node.js 12.x`, `Node.js 10.x`, `Python 3.8`, `Python 3.7`, `Ruby 2.7`, `Ruby 2.5`, `Custom runtime`                                                               |
| 0.0.1   | `Python 3.7`, `Python 3.8`                                                                                                                                                                                               |

**NOTE:** If your AWS region is not in the list, please reach out to Logz.io's support or open an issue in this repo.

### Changelog:
- **0.3.5**:
  - Fix ARM layer release
- **0.3.4**:
  - Fix missing Lambda layer policy in release.
- **0.3.3**:
  - Add `JSON_FIELDS_UNDER_ROOT` to allow nesting Json fields under the root instead of under `message_nested`
- **0.3.2**:
  - Bug fix for grok patterns - fields parsed incorrectly.
- **0.3.1**:
  - Remove ability to send extension logs.
- **0.3.0**:
  - Enrich logs with the following fields: `lambda_function_name`, `aws_region`.
  - Allow adding custom fields with `CUSTOM_FIELDS` env var.
- **0.2.0**:
  - Allow parsing log into fields. To learn more see [parsing logs](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#parsing-logs) section.
  - Allow nested JSON within logs. To learn more see [nested fields](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#nested-fields) section.
- **0.1.0**:
  - **BREAKING CHANGES**: Written in Go, supports multiple runtimes. Compatible with the GA version of the Extensions API.
- **0.0.1**: Initial release. Supports only python 3.7, python 3.8 runtimes. Compatible with the beta version of the Extensions API.
