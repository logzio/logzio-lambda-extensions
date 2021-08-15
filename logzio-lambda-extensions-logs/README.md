# Logzio-lambda-extensions-logs

Lambda extensions enable tools to integrate deeply into the Lambda execution environment to control and participate in Lambda’s lifecycle.
To read more about Lambda Extensions, [click here](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html).  
The Logz.io Lambda extension for logs, uses the AWS Extensions API and [AWS Logs API](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-logs-api.html), and sends your Lambda Function Logs directly to your Logz.io account.

This repo is based on the [AWS lambda extensions sample](https://github.com/aws-samples/aws-lambda-extensions).

This extension is written in Go, but can be run with runtimes that support extensions](https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html).

### Prerequisites

* Lambda function with [supported runtime](https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html) for extensions.
* AWS Lambda limitations: A function can use up to five layers at a time. The total unzipped size of the function and all layers cannot exceed the unzipped deployment package size limit of 250 MB.

### Important notes:
* If the extension won't have enough time to receive logs from AWS Logs API, it may send the logs in the next invocation of the Lambda function.
So if you want that all the logs will be sent by the end of your function's run, you'll need to add at the end of your Lambda function code a sleep interval that will allow the extension enough time to do it's job.
* Due to [Lambda's execution environment lifecycle](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-context.html), the extension is being invoked on two events - `INVOKE` and `SHUTDOWN`.
That means that if your Lambda function goes into the `SHUTDOWN` phase, the extension will run and if there are logs in it's queue, it will send them.

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

| Placeholder | Description |
|---|---|
| `<<FUNCTION-NAME>>` |  Name of the Lambda Function you want to monitor. |
| `<<LAYERS>>` | A space-separated list of function layers to add to the function's execution environment. Specify each layer by its ARN, including the version.  For the ARN, see the [**Lambda extension versions** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#lambda-extension-versions).|
| `<<ENV-VARS>>`  | Key-value pairs containing environment variables that are accessible from function code during execution. Should appear in the following format: `KeyName1=string,KeyName2=string`.  For a list of all the environment variables for the extension, see the [**Lambda environment variables** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#environment-variables).|

##### Run the function

Use the following command. It may take more than one run of the function for the logs to start shipping to your Logz.io account.

```shell
aws lambda update-function-configuration \
    --function-name <<FUNCTION-NAME>> \
    --layers [] \
    --environment "Variables={}"
```

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

3. Select the `Specify an ARN` option, then choose the ARN of the extension with the region code that matches your Lambda Function region from the [**Lambda extension versions** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#lambda-extension-versions), and click the `Add` button.
![Add ARN extension](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lambda_extensions/lambda-x_1-3.jpg)

##### Configure the extension parameters

Add the environment variables to the function, according to the [**Environment variables** table](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs#environment-variables).

##### Run the function

Run the function. It may take more than one run of the function for the logs to start shipping to your Logz.io account.

#### Deleting the extension

- To delete the **extension layer**: In your function page, go to the **layers** panel. Click `edit`, select the extension layer, and click `save`.
- To delete the extension's **environment variables**: In your function page, select the `Configuration` tab, select `Environment variables`, click `edit`, and remove the variables that you added for the extension.

### Environment Variables

| Name | Description |Required/Default|
| --- | --- | --- |
| `LOGZIO_LOGS_TOKEN` | Your Logz.io log shipping [token](https://app.logz.io/#/dashboard/settings/manage-tokens/data-shipping). | Required |
| `LOGZIO_LISTENER` |  Your  Logz.io listener address, with port 8070 (http) or 8071 (https). For example, for example: `https://listener.logz.io:8071` | Required |
| `LOGS_EXT_LOG_LEVEL` |  Log level of the extension. Can be set to one of the following: `debug`, `info`, `warn`, `error`, `fatal`, `panic`. |Default: `info` |
| `ENABLE_EXTENSION_LOGS` |  Set to `true` if you wish the extension logs will be shipped to your Logz.io account. | Default: `false` |
| `ENABLE_PLATFORM_LOGS` | The platform log captures runtime or execution environment errors. Set to `true` if you wish the platform logs will be shipped to your Logz.io account. | Default: `false` |

### Lambda extension versions

| Version | Supported Runtimes | AWS ARN |
| --- | --- | --- |
| 0.1.0| `.NET Core 3.1`, `Java 11`, `Java 8`, `Node.js 14.x`, `Node.js 12.x`, `Node.js 10.x`, `Python 3.8`, `Python 3.7`, `Ruby 2.7`, `Ruby 2.5`, `Custom runtime`| `arn:aws:lambda:<<YOUR-AWS-REGION-CODE>>:486140753397:layer:LogzioLambdaExtensionLogs:2` |
| 0.0.1 | `Python 3.7`, `Python 3.8` | `arn:aws:lambda:<<YOUR-AWS-REGION-CODE>>:486140753397:layer:LogzioLambdaExtensionLogs:1` |

### Available AWS regions

| Region Name | Region Code | Notes |
| --- | ---- | --- |
| US East (N. Virginia) | `us-east-1` | - |
| US East (Ohio) | `us-east-2` | - |
| US West (N. California) | `us-west-1` | - |
| US West (Oregon) | `us-west-2` | - |
| Europe (Frankfurt) | `eu-central-1`| - |
| Europe (Ireland) | `eu-west-1` | - |
| Europe (Stockholm) | `eu-north-1` | - |
| Asia Pacific (Sydney) | `ap-southeast-2` | Available from v0.1.0 |

**NOTE:** If your AWS region is not in the list, please reach out to Logz.io's support or open an issue in this repo.

### ARN for extension dependencies - DEPRECATED
Note: the dependencies will is deprecated.
|Compatible with extension versions | Imports | AWS ARN |
| --- | --- | --- |
| 0.0.1 | `requests` | `arn:aws:lambda:<<YOUR-AWS-REGION-CODE>>:486140753397:layer:LogzioLambdaExtensionLogsLibs:1` |

### Changelog:
- **0.1.0**:
    - **BREAKING CHANGES**: Written in Go, supports multiple runtimes. Compatible with the GA version of the Extensions API.
- **0.0.1**: Initial release. Supports only python 3.7, python 3.8 runtimes. Compatible with the beta version of the Extensions API.
