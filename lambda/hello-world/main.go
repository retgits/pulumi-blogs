package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	val := os.Getenv("NAME")
	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %s", val),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
