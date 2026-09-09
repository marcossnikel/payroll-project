package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type PayrollStackProps struct {
	awscdk.StackProps
}

func newPayrollStack(scope constructs.Construct, id string, props *PayrollStackProps) awscdk.Stack {
	var stackProps awscdk.StackProps
	if props != nil {
		stackProps = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &stackProps)

	table := awsdynamodb.NewTable(stack, jsii.String("PayrollData"), &awsdynamodb.TableProps{
		PartitionKey:  &awsdynamodb.Attribute{Name: jsii.String("PK"), Type: awsdynamodb.AttributeType_STRING},
		SortKey:       &awsdynamodb.Attribute{Name: jsii.String("SK"), Type: awsdynamodb.AttributeType_STRING},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		Stream:        awsdynamodb.StreamViewType_NEW_AND_OLD_IMAGES,
		Encryption:    awsdynamodb.TableEncryption_AWS_MANAGED,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	effectDLQ := queue(stack, "PaymentEffectDLQ", nil)
	auditDLQ := queue(stack, "AuditDLQ", nil)
	effectQueue := queue(stack, "PaymentEffectQueue", effectDLQ)
	auditQueue := queue(stack, "AuditQueue", auditDLQ)
	commands := awssns.NewTopic(stack, jsii.String("PaymentCommands"), &awssns.TopicProps{
		DisplayName: jsii.String("Synthetic payroll payment commands"),
	})
	commands.AddSubscription(awssnssubscriptions.NewSqsSubscription(effectQueue, nil))
	commands.AddSubscription(awssnssubscriptions.NewSqsSubscription(auditQueue, nil))

	apiFunction := goFunction(stack, "API", "../build/lambda/api", map[string]*string{
		"TABLE_NAME": table.TableName(),
	})
	effectFunction := goFunction(stack, "PaymentEffectWorker", "../build/lambda/effect-worker", map[string]*string{
		"TABLE_NAME": table.TableName(),
	})
	outboxFunction := goFunction(stack, "OutboxPublisher", "../build/lambda/outbox-publisher", map[string]*string{
		"TOPIC_ARN": commands.TopicArn(),
	})
	reconcilerFunction := goFunction(stack, "Reconciler", "../build/lambda/reconciler", map[string]*string{
		"TABLE_NAME": table.TableName(),
	})

	table.GrantReadWriteData(apiFunction)
	table.GrantReadWriteData(effectFunction)
	table.GrantReadWriteData(reconcilerFunction)
	table.GrantStreamRead(outboxFunction)
	commands.GrantPublish(outboxFunction)

	effectFunction.AddEventSource(awslambdaeventsources.NewSqsEventSource(effectQueue, &awslambdaeventsources.SqsEventSourceProps{
		BatchSize: jsii.Number(5),
	}))
	outboxFunction.AddEventSource(awslambdaeventsources.NewDynamoEventSource(table, &awslambdaeventsources.DynamoEventSourceProps{
		StartingPosition: awslambda.StartingPosition_LATEST,
		BatchSize:        jsii.Number(25),
		RetryAttempts:    jsii.Number(3),
	}))

	httpAPI := awsapigatewayv2.NewHttpApi(stack, jsii.String("PayrollHTTPAPI"), &awsapigatewayv2.HttpApiProps{
		ApiName:     jsii.String("payroll-project"),
		Description: jsii.String("Synthetic payroll operations API"),
	})
	httpAPI.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/{proxy+}"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_ANY},
		Integration: awsapigatewayv2integrations.NewHttpLambdaIntegration(jsii.String("APILambdaIntegration"), apiFunction, nil),
	})

	schedule := awsevents.NewRule(stack, jsii.String("ReconciliationSchedule"), &awsevents.RuleProps{
		Description: jsii.String("Find and reconcile ambiguous worker payments"),
		Schedule:    awsevents.Schedule_Rate(awscdk.Duration_Minutes(jsii.Number(5))),
	})
	schedule.AddTarget(awseventstargets.NewLambdaFunction(reconcilerFunction, nil))

	addDLQAlarm(stack, "PaymentEffectDLQAlarm", effectDLQ)
	addDLQAlarm(stack, "AuditDLQAlarm", auditDLQ)

	awscdk.NewCfnOutput(stack, jsii.String("APIURL"), &awscdk.CfnOutputProps{
		Value:       httpAPI.ApiEndpoint(),
		Description: jsii.String("HTTP API endpoint"),
	})
	return stack
}

func queue(scope constructs.Construct, id string, deadLetterQueue awssqs.IQueue) awssqs.Queue {
	props := &awssqs.QueueProps{
		Encryption:        awssqs.QueueEncryption_SQS_MANAGED,
		RetentionPeriod:   awscdk.Duration_Days(jsii.Number(14)),
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(60)),
	}
	if deadLetterQueue != nil {
		props.DeadLetterQueue = &awssqs.DeadLetterQueue{
			Queue:           deadLetterQueue,
			MaxReceiveCount: jsii.Number(5),
		}
	}
	return awssqs.NewQueue(scope, &id, props)
}

func goFunction(scope constructs.Construct, id, assetPath string, environment map[string]*string) awslambda.Function {
	return awslambda.NewFunction(scope, &id, &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(&assetPath, nil),
		MemorySize:   jsii.Number(256),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Environment:  &environment,
		Tracing:      awslambda.Tracing_ACTIVE,
	})
}

func addDLQAlarm(scope constructs.Construct, id string, queue awssqs.IQueue) {
	awscloudwatch.NewAlarm(scope, &id, &awscloudwatch.AlarmProps{
		Metric:             queue.MetricApproximateNumberOfMessagesVisible(nil),
		Threshold:          jsii.Number(0),
		EvaluationPeriods:  jsii.Number(1),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
		TreatMissingData:   awscloudwatch.TreatMissingData_NOT_BREACHING,
		AlarmDescription:   jsii.String("Synthetic payroll messages reached a dead-letter queue"),
	})
}

func main() {
	defer jsii.Close()
	app := awscdk.NewApp(nil)
	newPayrollStack(app, "PayrollProjectStack", &PayrollStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})
	app.Synth(nil)
}

func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")
	if account == "" || region == "" {
		return nil
	}
	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
