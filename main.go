package main

import (
	"fmt"
	"strings"

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

		// Create the required number of subnets
		subnets := make(map[string]interface{})
		subnets["subnet_ids"] = make([]interface{}, 0)

		subnetZones := strings.Split(getEnv(ctx, "vpc:subnet-zones", "unknown"), ",")
		subnetIPs := strings.Split(getEnv(ctx, "vpc:subnet-ips", "unknown"), ",")

		for idx, availabilityZone := range subnetZones {
			subnetArgs := &ec2.SubnetArgs{
				Tags:             tags,
				VpcId:            vpc.ID(),
				CidrBlock:        subnetIPs[idx],
				AvailabilityZone: availabilityZone,
			}

			subnet, err := ec2.NewSubnet(ctx, fmt.Sprintf("%s-subnet-%d", vpcName, idx), subnetArgs)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			subnets["subnet_ids"] = append(subnets["subnet_ids"].([]interface{}), subnet.ID())
		}

		ctx.Export("SUBNET-IDS", subnets["subnet_ids"])

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
