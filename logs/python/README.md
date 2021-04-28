# Logzio-lambda-extensions

Lambda extensions enable tools to integrate deeply into the Lambda execution environment to control and participate in Lambda’s lifecycle.
To read more about Lambda Extensions, [click here](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html).  
The Logz.io Lambda extension for logs, uses the AWS Extensions API and [AWS Logs API](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-logs-api.html), and sends your Lambda Function Logs directly to your Logz.io account.

This repo is based on the [AWS lambda extensions sample](https://github.com/aws-samples/aws-lambda-extensions/tree/main/python-example-logs-api-extension/extensions).

### Prerequisites

* Lambda function with supported runtime for extensions (`python 3.7`, `python 3.8`).
* AWS Lambda limitations: A function can use up to five layers at a time. The total unzipped size of the function and all layers cannot exceed the unzipped deployment package size limit of 250 MB.

### Extension deployment options

You can deploy the extension via:
* [AWS CLI]().
* [AWS Management Console]().

#### Deploying Logz.io logs extension via the AWS CLI

##### Deploy the extension, dependencies and configuration

If you haven't done it already, [install](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) and [configure](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) the AWS CLI.

Add the layer to your function and configure the environment variables using the following command:

```shell
aws lambda update-function-configuration \
    --function-name <<FUNCTION-NAME>> \
    --layers <<LAYERS>> \
    --environment "Variables={<<ENV-VARS>>}"
```

**Note:** this command overwrites the existing function configuration. If you already have your own layers and environment variables for your function, list them as well.

| Placeholder | Description | Required/Default|
|---|---|---|
| `<<FUNCTION-NAME>>` |  Name of the Lambda Function you want to monitor. |Required|
| `<<LAYERS>>` | A space-separated list of function layers to add to the function's execution environment. Specify each layer by its ARN, including the version.  For the ARN, see the [**Lambda extension versions** table](). You may need to add another layer that has the extensions dependencies.  For the libraries that your extension requires, see the [**Lambda extensions dependencies** table](). If your function doesn't already have those libraries under `/opt/python`, you'll need to add the dependencies as a layer, too.|  |
| `<<ENV-VARS>>`  | Key-value pairs containing environment variables that are accessible from function code during execution. Should appear in the following format: `KeyName1=string,KeyName2=string`.  For a list of all the environment variables for the extension, see the [**Lambda environment variables** table]().|  |

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

To delete the extension, its dependencies and environment variables, use the following command:

```shell
aws lambda update-function-configuration \
    --function-name some-func \
    --layers [] \
    --environment "Variables={}"
```

**NOTE:** This command overwrites the existing function configuration. If you already have your own layers and environment variables for your function, include them in the list.

#### Deploying Logz.io log extensions via the AWS Management Console

##### Add the extension to your Lambda Function

1. In the Lambda Functions screen, choose the function you want to monitor.
![Pick lambda function](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lambda_extensions/lambda-x_1-1.jpg)

2. In the page for the function, scroll down to the `Layers` section and choose `Add Layer`.
![Add layer](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lambda_extensions/lambda-x_1-2.jpg)

3. Select the `Specify an ARN` option, then choose the ARN of the extension with the region code that matches your Lambda Function region from the [**Lambda extension versions** table](), and click the `Add` button.
![Add ARN extension](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lambda_extensions/lambda-x_1-3.jpg)

4. *Optional*. This step adds the python libraries the extension needs to run. Refer to the [**ARN for extension dependencies** table](): If your Lambda function already has those libraries under `/opt/python`, you can skip this step. If not, you'll need it for the extension to run.

    a. Repeat step 2 to add another layer.
  
    b. Select the `Specify an ARN` option, then select the ARN that's compatible with the extension version you chose from the [**Dependencies** table](), and paste it in the textbox. 
  
    c. Click `Add`.

##### Configure the extension parameters

Add the environment variables to the function, according to the [**Environment variables** table]().

##### Run the function

Run the function. It may take more than one run of the function for the logs to start shipping to your Logz.io account.

#### Deleting the extension

- To delete the **extension layer**: In your function page, go to the **layers** panel. Click `edit`, select the extension layer, and click `save`.
- To delete the **extension dependencies layer**: In your function page, go to the **layers** panel, click `edit`, select the extension dependencies layer, and click `save`.
- To delete the extension's **environment variables**: In your function page, select the `Configuration` tab, select `Environment variables`, click `edit`, and remove the variables that you added for the extension.

### Environment Variables

| Name | Description |Required/Default|
| --- | --- | --- |
| `LOGZIO_LOGS_TOKEN` | Your Logz.io log shipping [token](https://app.logz.io/#/dashboard/settings/manage-tokens/data-shipping). | Required |
| `LOGZIO_REGION` |  Two-letter region code, or blank for US. This determines your listener URL. You can find your region code in the [Regions and URLs table](https://docs.logz.io/user-guide/accounts/account-region.html#regions-and-urls). | Default: (left blank for US)|
| `LOGS_EXT_LOG_LEVEL` |  Log level of the extension. Can be set to one of the following: `DEBUG`, `INFO`, `WARNING`, `ERROR`, `CRITICAL`. |Default: `INFO` |
| `ENABLE_EXTENSION_LOGS` |  Set to `true` if you wish the extension logs will be shipped to your Logz.io account. | Default: `false` |
| `ENABLE_PLATFORM_LOGS` | The platform log captures runtime or execution environment errors. Set to `true` if you wish the platform logs will be shipped to your Logz.io account. | Default: `false` |
| `THREAD_TIMEOUT` | Execution timeout for the threads that parse and send the logs. | Default: `5` (seconds)|
| `LOGZIO_CUSTOM_LISTENER` | Use if you have a custom Logz.io listener endpoint). Will override `LOGZIO_REGION`. | Optional |


### Lambda extension versions

| Version | Supported Runtimes | AWS ARN |
| --- | --- | --- |
| 0.0.1 | python 3.7, python 3.8 | `arn:aws:lambda:<<YOUR-AWS-REGION-CODE>>:486140753397:layer:LogzioLambdaExtensionLogs:1` |

### ARN for extension dependencies

|Compatible with extension versions | Imports | AWS ARN |
| --- | --- | --- |
| 0.0.1 | `requests` | `arn:aws:lambda:<<YOUR-AWS-REGION-CODE>>:486140753397:layer:LogzioLambdaExtensionLogsLibs:1` |

### Available AWS regions

| Region Name | Region Code |
| --- | ---- |
| US East (N. Virginia) | `us-east-1` |
| US East (Ohio) | `us-east-2` |
| US West (N. California) | `us-west-1` |
| US West (Oregon) | `us-west-2` |
| Europe (Frankfurt) | `eu-central-1` |
| Europe (Ireland) | `eu-west-1` |
| Europe (Stockholm) | `eu-north-1` |

**NOTE:** If your AWS region is not in the list, please reach out to Logz.io's support or open an issue in this repo.