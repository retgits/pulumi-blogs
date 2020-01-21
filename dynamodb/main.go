package main

import (
	"github.com/pulumi/pulumi-aws/sdk/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

// DynamoAttribute represents an attribute for describing the key schema for the table and indexes.
type DynamoAttribute struct {
	Name string
	Type string
}

// DynamoAttributes is an array of DynamoAttribute
type DynamoAttributes []DynamoAttribute

// ToList takes a DynamoAttributes object and turns that into a slice of map[string]interface{} so it can be correctly passed to the Pulumi runtime
func (d DynamoAttributes) ToList() []map[string]interface{} {
	array := make([]map[string]interface{}, len(d))
	for idx, attr := range d {
		m := make(map[string]interface{})
		m["name"] = attr.Name
		m["type"] = attr.Type
		array[idx] = m
	}
	return array
}

// GlobalSecondaryIndex represents the properties of a global secondary index
type GlobalSecondaryIndex struct {
	Name           string
	HashKey        string
	ProjectionType string
	WriteCapacity  int
	ReadCapacity   int
}

// GlobalSecondaryIndexes is an array of GlobalSecondaryIndex
type GlobalSecondaryIndexes []GlobalSecondaryIndex

// ToList takes a GlobalSecondaryIndexes object and turns that into a slice of map[string]interface{} so it can be correctly passed to the Pulumi runtime
func (g GlobalSecondaryIndexes) ToList() []map[string]interface{} {
	array := make([]map[string]interface{}, len(g))
	for idx, attr := range g {
		m := make(map[string]interface{})
		m["name"] = attr.Name
		m["hash_key"] = attr.HashKey
		m["projection_type"] = attr.ProjectionType
		m["write_capacity"] = attr.WriteCapacity
		m["read_capacity"] = attr.ReadCapacity
		array[idx] = m
	}
	return array
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create the attributes for ID and User
		dynamoAttributes := DynamoAttributes{
			DynamoAttribute{
				Name: "ID",
				Type: "S",
			},
			DynamoAttribute{
				Name: "User",
				Type: "S",
			},
		}

		// Create a Global Secondary Index for the user field
		gsi := GlobalSecondaryIndexes{
			GlobalSecondaryIndex{
				Name: "User",
				HashKey: "User",
				ProjectionType: "ALL",
				WriteCapacity: 10,
				ReadCapacity: 10,
			},
		}

		// Create a TableArgs struct that contains all the data
		tableArgs := &dynamodb.TableArgs{
			Attributes:    dynamoAttributes.ToList(),
			HashKey:       "ID",
			WriteCapacity: 10,
			ReadCapacity:  10,
			GlobalSecondaryIndexes: gsi.ToList(),
		}

		// Let the Pulumi runtime create the table
		userTable, err := dynamodb.NewTable(ctx, "User", tableArgs)
		if err != nil {
			return err
		}

		// Export the name of the newly created table as an output in the stack
		ctx.Export("TableName", userTable.ID())
	})
}