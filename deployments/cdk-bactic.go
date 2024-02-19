package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkDemoStackProps struct {
	awscdk.StackProps
}

func NewBacticStack(scope constructs.Construct, id string, props *CdkDemoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	// For now, just get the default VPC in our region
	// vpc := awsec2.Vpc_FromLookup(stack, jsii.String("VPC"), &awsec2.VpcLookupOptions{
	// 	IsDefault: jsii.Bool(true),
	// })

	// awsec2.NewPublicSubnet(
	// 	stack, jsii.String("bactic-public-subnet"), &awsec2.PublicSubnetProps{
	// 		AvailabilityZone:    jsii.String("us-west-1a"),
	// 		CidrBlock:           jsii.String("10.0.1.0/24"),
	// 		VpcId:               vpc.VpcId(),
	// 		MapPublicIpOnLaunch: jsii.Bool(true),
	// 	})

	// awsec2.NewPrivateSubnet(stack, jsii.String("bactic-private-subnet"), &awsec2.PrivateSubnetProps{
	// 	AvailabilityZone:    jsii.String("us-west-1a"),
	// 	CidrBlock:           jsii.String("10.0.2.0/24"),
	// 	VpcId:               vpc.VpcId(),
	// 	MapPublicIpOnLaunch: jsii.Bool(false),
	// })

	cluster := awsecs.NewCluster(stack, jsii.String("bactic-cluster"), &awsecs.ClusterProps{})

	// Database setup
	dbTask := awsecs.NewTaskDefinition(stack, jsii.String("DatabaseTask"), &awsecs.TaskDefinitionProps{
		Compatibility: awsecs.Compatibility_FARGATE,
		Cpu:           jsii.String("256"),
		MemoryMiB:     jsii.String("512"),
	})

	postgres_pass := "tmp_password"

	dbContainer := dbTask.AddContainer(jsii.String("postgres-bactic"), &awsecs.ContainerDefinitionOptions{
		Image:          awsecs.AssetImage_FromRegistry(jsii.String("postgres:15.4"), nil),
		MemoryLimitMiB: jsii.Number(512),
	})
	dbContainer.AddEnvironment(jsii.String("POSTGRES_PASSWORD"), jsii.String(postgres_pass))
	dbContainer.AddEnvironment(jsii.String("POSTGRES_DB"), jsii.String("bactic"))
	dbContainer.AddPortMappings(&awsecs.PortMapping{
		ContainerPort: jsii.Number(5432),
		HostPort:      jsii.Number(5432),
	})

	awsecs.NewFargateService(stack, jsii.String("bactic-database"), &awsecs.FargateServiceProps{
		Cluster:        cluster,
		TaskDefinition: dbTask,
	})

	// Define scraper task
	scraperTask := awsecs.NewTaskDefinition(stack, jsii.String("bactic-scraper-task"), &awsecs.TaskDefinitionProps{
		Compatibility: awsecs.Compatibility_FARGATE,
		Cpu:           jsii.String("256"),
		MemoryMiB:     jsii.String("512"),
	})

	scraperContainer := scraperTask.AddContainer(jsii.String("bactic-scraper-container"), &awsecs.ContainerDefinitionOptions{
		Image: awsecs.ContainerImage_FromDockerImageAsset(awsecrassets.NewDockerImageAsset(stack, jsii.String("bactic-scraper-image"), &awsecrassets.DockerImageAssetProps{
			File:      jsii.String("Dockerfile.scraper"),
			AssetName: jsii.String("Scraper"),
			Directory: jsii.String("."),
		})),
		MemoryLimitMiB: jsii.Number(512),
	})
	scraperContainer.AddEnvironment(jsii.String("DB_PASS"), jsii.String(postgres_pass))

	// Define web task

	webTask := awsecs.NewTaskDefinition(stack, jsii.String("bactic-web-task"), &awsecs.TaskDefinitionProps{
		Compatibility: awsecs.Compatibility_FARGATE,
		Cpu:           jsii.String("256"),
		MemoryMiB:     jsii.String("512"),
	})

	webTask.AddContainer(jsii.String("bactic-web-container"), &awsecs.ContainerDefinitionOptions{
		Image: awsecs.ContainerImage_FromDockerImageAsset(awsecrassets.NewDockerImageAsset(stack, jsii.String("WebServerImage"), &awsecrassets.DockerImageAssetProps{
			File:      jsii.String("Dockerfile.web"),
			AssetName: jsii.String("Web"),
			Directory: jsii.String("."),
		})),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewBacticStack(app, "BacticStack", &CdkDemoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String("us-west-1"),
	}
}
