package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dbItem map[string]string

type verifyMoviesEvent struct {
	SourceName string `json:"sourceName"`
	BatchID    string `json:"batchID"`
	BatchDate  string `json:"batchDate"`
	TotalItems int64  `json:"totalItems"`
}

var moviesTableName = "movies"
var timeInSecondsWaitToFinishImportProcess = 2
var attemptQuantity = 5

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event verifyMoviesEvent) error {
	fmt.Printf("handler > sourceName:%s batchID:%s BatchDate:%s totalItems:%d\n", event.SourceName, event.BatchID, event.BatchDate, event.TotalItems)

	success := false

	var currentCount int64

	for i := 0; i < attemptQuantity; i++ {

		time.Sleep(time.Duration(timeInSecondsWaitToFinishImportProcess) * time.Second)

		currentCount = processCountByBatchID(event.BatchID)

		if currentCount == event.TotalItems {
			fmt.Println("process finished with success > batchID:", event.BatchID, "batchDate:", event.BatchDate, "totalItems:", event.TotalItems)
			success = true
		} else {
			fmt.Println("ongoing process > batchID:", event.BatchID, "batchDate:", event.BatchDate, "totalItems:", event.TotalItems, "currentCount:", currentCount)
		}
	}

	if success == false {
		return errors.New(fmt.Sprintln("process finished with error > batchID:", event.BatchID, "batchDate:", event.BatchDate, "totalItems:", event.TotalItems, "currentCount:", currentCount))
	}

	return nil
}

func processCountByBatchID(batchID string) int64 {

	db := dynamodb.New(session.New())

	var selctCount = dynamodb.SelectCount
	var countQuery = map[string]*dynamodb.Condition{
		"batchID": {
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{
				{
					N: aws.String(batchID),
				},
			},
		},
	}

	countItemsInMoviesTableInput := dynamodb.QueryInput{
		TableName:     &moviesTableName,
		Select:        &selctCount,
		KeyConditions: countQuery,
	}

	countItemsInMoviesTableOutput, err := db.Query(&countItemsInMoviesTableInput)
	if err != nil {
		log.Fatal(err)
	}

	return *countItemsInMoviesTableOutput.Count
}
