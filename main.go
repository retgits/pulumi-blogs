package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Prepare the tags that are used for each individual resource so they can be found
		// using the Resource Groups service in the AWS Console
		tags := make(map[string]interface{})
		tags["version"] = getEnv(ctx, "tags:version", "unknown")
		tags["author"] = getEnv(ctx, "tags:author", "unknown")
		tags["team"] = getEnv(ctx, "tags:team", "unknown")
		tags["feature"] = getEnv(ctx, "tags:feature", "unknown")
		tags["region"] = getEnv(ctx, "aws:region", "unknown")

		// Create a VPC for the EKS cluster
		cidrBlock := getEnv(ctx, "vpc:cidr-block", "unknown")

		vpcArgs := &ec2.VpcArgs{
			CidrBlock: cidrBlock,
			Tags:      tags,
		}

		vpcName := getEnv(ctx, "vpc:name", "unknown")

		vpc, err := ec2.NewVpc(ctx, vpcName, vpcArgs)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		// Export IDs of the created resources to the Pulumi stack
		ctx.Export("VPC-ID", vpc.ID())

		return nil
	})
}

// getEnv searches for the requested key in the pulumi context and provides either the value of the key or the fallback.
func getEnv(ctx *pulumi.Context, key string, fallback string) string {
	if value, ok := ctx.GetConfig(key); ok {
		return value
	}
	return fallback
}
