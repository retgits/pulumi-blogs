package main

import (
	"github.com/pulumi/pulumi-aws/sdk/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create the attributes for ID and User
		dynamoAttributes := []dynamodb.TableAttributeInput{
			dynamodb.TableAttributeArgs{
				Name: pulumi.String("ID"),
				Type: pulumi.String("S"),
			}, dynamodb.TableAttributeArgs{
				Name: pulumi.String("User"),
				Type: pulumi.String("S"),
			},
		}

		// Create a Global Secondary Index for the user field
		gsi := []dynamodb.TableGlobalSecondaryIndexInput{
			dynamodb.TableGlobalSecondaryIndexArgs{
				Name:           pulumi.String("User"),
				HashKey:        pulumi.String("User"),
				ProjectionType: pulumi.String("ALL"),
				WriteCapacity:  pulumi.Int(10),
				ReadCapacity:   pulumi.Int(10),
			},
		}

		// Create a TableArgs struct that contains all the data
		tableArgs := &dynamodb.TableArgs{
			Attributes:             dynamodb.TableAttributeArray(dynamoAttributes),
			HashKey:                pulumi.String("ID"),
			WriteCapacity:          pulumi.Int(10),
			ReadCapacity:           pulumi.Int(10),
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray(gsi),
		}

		// Let the Pulumi runtime create the table
		userTable, err := dynamodb.NewTable(ctx, "User", tableArgs)
		if err != nil {
			return err
		}

		// Export the name of the newly created table as an output in the stack
		ctx.Export("TableName", userTable.ID())

		return nil
	})
}
