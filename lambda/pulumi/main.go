package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

const (
	shell      = "sh"
	shellFlag  = "-c"
	rootFolder = "/rootfolder/of/your/lambdaapp"
)

func runCmd(args string) error {
	cmd := exec.Command(shell, shellFlag, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = rootFolder
	return cmd.Run()
}

func main() {
	if err := runCmd("GOOS=linux GOARCH=amd64 go build -o hello-world/hello-world ./hello-world"); err != nil {
		fmt.Printf("Error building code: %s", err.Error())
		os.Exit(1)
	}

	if err := runCmd("zip ./hello-world/hello-world.zip ./hello-world/hello-world"); err != nil {
		fmt.Printf("Error creating zipfile: %s", err.Error())
		os.Exit(1)
	}

	if err := runCmd("aws s3 cp ./hello-world/hello-world.zip s3://us-west-2-retgits-lambda-apps/hello-world.zip"); err != nil {
		fmt.Printf("Error creating zipfile: %s", err.Error())
		os.Exit(1)
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		// The policy description of the IAM role, in this case only the sts:AssumeRole is needed
		roleArgs := &iam.RoleArgs{
			AssumeRolePolicy: `{
        "Version": "2012-10-17",
        "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "lambda.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
        ]
    }`,
		}

		// Create a new role called HelloWorldIAMRole
		role, err := iam.NewRole(ctx, "HelloWorldIAMRole", roleArgs)
		if err != nil {
			fmt.Printf("role error: %s\n", err.Error())
			return err
		}

		// Export the role ARN as an output of the Pulumi stack
		ctx.Export("Role ARN", role.Arn())

		environment := make(map[string]interface{})
		variables := make(map[string]interface{})
		variables["NAME"] = "WORLD"
		environment["variables"] = variables

		// The set of arguments for constructing a Function resource.
		functionArgs := &lambda.FunctionArgs{
			Description: "My Lambda function",
			Runtime:     "go1.x",
			Name:        "HelloWorldFunction",
			MemorySize:  256,
			Timeout:     10,
			Handler:     "hello-world",
			Environment: environment,
			S3Bucket:    "us-west-2-retgits-lambda-apps",
			S3Key:       "hello-world.zip",
			Role:        role.Arn(),
		}

		// NewFunction registers a new resource with the given unique name, arguments, and options.
		function, err := lambda.NewFunction(ctx, "HelloWorldFunction", functionArgs)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		// Export the function ARN as an output of the Pulumi stack
		ctx.Export("Function", function.Arn())

		return nil
	})
}
