package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/szazeski/checkssl/lib/checkssl"
	"time"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	dateThreshold := time.Now()
	result := checkssl.CheckServer(request.QueryStringParameters["target"], dateThreshold, false)

	headers := map[string]string{"Access-Control-Allow-Origin": "*", "Access-Control-Allow-Methods": "GET,OPTIONS", "Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token"}

	if result.ExitCode == 0 {
		return events.APIGatewayProxyResponse{Body: result.AsString(false), Headers: headers, StatusCode: 200}, nil
	} else {
		return events.APIGatewayProxyResponse{Body: result.AsString(false), Headers: headers, StatusCode: 401}, nil
	}
}

func main() {
	lambda.Start(Handler)
}
