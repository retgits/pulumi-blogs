package main

import (
	"fmt"
	"log"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

type EKSConfig struct {
	ClusterLogTypes   []string `json:"cluster-log-types"`
	ClusterName       string   `json:"cluster-name"`
	ClusterRoleARN    string   `json:"cluster-role-arn"`
	KubernetesVersion string   `json:"k8s-version"`
}

type FargateConfig struct {
	ExecutionRoleARN string `json:"execution-role-arn"`
	Namespace        string `json:"namespace"`
	ProfileName      string `json:"profile-name"`
}

type Tags struct {
	Author  string
	Feature string
	Team    string
	Version string
	Stage   string
}

type VPCConfig struct {
	CIDRBlock   string `json:"cidr-block"`
	Name        string
	SubnetIPs   []string `json:"subnet-ips"`
	SubnetZones []string `json:"subnet-zones"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a new config object with the data from the YAML file
		// The object has all the data that the namespace awsconfig has
		conf := config.New(ctx, "awsconfig")

		// Prepare the tags that are used for each individual resource so they can be found
		// using the Resource Groups service in the AWS Console
		var tags Tags
		conf.RequireObject("tags", &tags)
		tagMap := make(map[string]pulumi.Input)
		tagMap["author"] = pulumi.String(tags.Author)
		tagMap["team"] = pulumi.String(tags.Team)
		tagMap["version"] = pulumi.String(tags.Version)
		tagMap["feature"] = pulumi.String(tags.Feature)
		tagMap["stage"] = pulumi.String(tags.Stage)

		// Create a VPC for the EKS cluster
		var vpcConfig VPCConfig
		conf.RequireObject("vpc", &vpcConfig)

		vpcArgs := &ec2.VpcArgs{
			CidrBlock: pulumi.String(vpcConfig.CIDRBlock),
			Tags:      pulumi.Map(tagMap),
		}

		vpc, err := ec2.NewVpc(ctx, vpcConfig.Name, vpcArgs)
		if err != nil {
			log.Printf("error creating VPC: %s", err.Error())
			return err
		}

		// Export IDs of the created resources to the Pulumi stack
		ctx.Export("VPC-ID", vpc.ID())

		// Create the required number of subnets
		subnets := make([]pulumi.StringInput, len(vpcConfig.SubnetZones))
		for idx, availabilityZone := range vpcConfig.SubnetZones {
			subnetArgs := &ec2.SubnetArgs{
				Tags:             pulumi.Map(tagMap),
				VpcId:            vpc.ID(),
				CidrBlock:        pulumi.String(vpcConfig.SubnetIPs[idx]),
				AvailabilityZone: pulumi.String(availabilityZone),
			}

			subnet, err := ec2.NewSubnet(ctx, fmt.Sprintf("%s-subnet-%d", vpcConfig.Name, idx), subnetArgs)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			subnets = append(subnets, subnet.ID())
		}

		// Create an EKS cluster
		var eksConfig EKSConfig
		conf.RequireObject("eks", &eksConfig)

		logTypes := make([]pulumi.StringInput, len(eksConfig.ClusterLogTypes))
		for idx, val := range eksConfig.ClusterLogTypes {
			logTypes[idx] = pulumi.String(val)
		}

		clusterArgs := &eks.ClusterArgs{
			Name:    pulumi.String(eksConfig.ClusterName),
			Version: pulumi.String(eksConfig.KubernetesVersion),
			RoleArn: pulumi.String(eksConfig.ClusterRoleARN),
			Tags:    pulumi.Map(tagMap),
			VpcConfig: eks.ClusterVpcConfigArgs{
				VpcId:     vpc.ID(),
				SubnetIds: pulumi.StringArray(subnets),
			},
			EnabledClusterLogTypes: pulumi.StringArray(logTypes),
		}

		cluster, err := eks.NewCluster(ctx, eksConfig.ClusterName, clusterArgs)
		if err != nil {
			log.Printf("error creating EKS cluster: %s", err.Error())
			return err
		}

		ctx.Export("CLUSTER-ID", cluster.ID())

		// Create an EKS Fargate Profile
		var fargateConfig FargateConfig
		conf.RequireObject("fargate", &fargateConfig)

		selectors := make([]eks.FargateProfileSelectorInput, 1)
		selectors[0] = eks.FargateProfileSelectorArgs{
			Namespace: pulumi.String(fargateConfig.Namespace),
		}

		fargateProfileArgs := &eks.FargateProfileArgs{
			ClusterName:         pulumi.String(eksConfig.ClusterName),
			FargateProfileName:  pulumi.String(fargateConfig.ProfileName),
			Tags:                pulumi.Map(tagMap),
			SubnetIds:           pulumi.StringArray(subnets),
			Selectors:           eks.FargateProfileSelectorArray(selectors),
			PodExecutionRoleArn: pulumi.String(fargateConfig.ExecutionRoleARN),
		}

		fargateProfile, err := eks.NewFargateProfile(ctx, fargateConfig.ProfileName, fargateProfileArgs)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		ctx.Export("FARGATE-PROFILE-ID", fargateProfile.ID())

		return nil
	})
}
