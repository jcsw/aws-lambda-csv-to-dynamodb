package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dbItem map[string]string

type createMoviesEvent struct {
	SourceName string   `json:"sourceName"`
	Items      []dbItem `json:"items"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event createMoviesEvent) error {
	fmt.Printf("handler > sourceName:%s itemsSize:%d\n", event.SourceName, len(event.Items))

	var db = dynamodb.New(session.New())

	for _, item := range event.Items {
		putItemInput := makeDynamodbItem(item)
		_, err := db.PutItem(&putItemInput)

		if err != nil {
			fmt.Println("handler > err", err)
		}
	}

	return nil
}

func makeDynamodbItem(item dbItem) dynamodb.PutItemInput {
	input := dynamodb.PutItemInput{
		TableName: aws.String("movies"),
		Item: map[string]*dynamodb.AttributeValue{
			"imdb": {
				S: aws.String(item["imdb"]),
			},
			"year": {
				N: aws.String(item["year"]),
			},
			"title": {
				S: aws.String(item["title"]),
			},
			"code": {
				S: aws.String(item["code"]),
			},
		},
	}

	return input
}
