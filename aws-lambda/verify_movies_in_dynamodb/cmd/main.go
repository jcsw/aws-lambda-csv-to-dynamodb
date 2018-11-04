package main

import (
	"context"
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
var timeInSecondsWaitToFinishImportProcess = 60
var attemptQuantity = 7

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event verifyMoviesEvent) {
	fmt.Printf("handler > sourceName:%s batchID:%s batchDate:%s totalItems:%d\n", event.SourceName, event.BatchID, event.BatchDate, event.TotalItems)

	success := false

	var currentCount int64

	for i := 0; i < attemptQuantity; i++ {

		time.Sleep(time.Duration(timeInSecondsWaitToFinishImportProcess) * time.Second)

		currentCount = processCountByBatchID(event.BatchID)

		if currentCount == event.TotalItems {
			fmt.Printf("process finished with success > batchID=%s batchDate=%s totalItems=%d totalImported:%d\n",
				event.BatchID, event.BatchDate, event.TotalItems, currentCount)
			success = true
			break
		} else {
			fmt.Printf("ongoing process > batchID=%s batchDate=%s totalItems=%d totalImported:%d\n",
				event.BatchID, event.BatchDate, event.TotalItems, currentCount)
		}
	}

	updateTableMoviesWriteThroughput()

	if success == false {
		fmt.Printf("process finished with error > batchID=%s batchDate=%s totalItems=%d totalImported:%d",
			event.BatchID, event.BatchDate, event.TotalItems, currentCount)
	}
}

func processCountByBatchID(batchID string) int64 {

	db := dynamodb.New(session.New())

	var selctCount = dynamodb.SelectCount
	var consistentRead = true

	countInput := dynamodb.QueryInput{
		TableName: aws.String(moviesTableName),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":batchID": {
				N: aws.String(batchID),
			},
		},
		KeyConditionExpression: aws.String("batchID = :batchID"),
		Select:                 &selctCount,
		ConsistentRead:         &consistentRead,
	}

	count := int64(0)

	countOutput, err := db.Query(&countInput)
	if err != nil {
		log.Fatal(err)
	}

	count += *countOutput.Count
	for len(countOutput.LastEvaluatedKey) > 0 {

		countInput.SetExclusiveStartKey(countOutput.LastEvaluatedKey)
		countOutput, err = db.Query(&countInput)
		if err != nil {
			log.Fatal(err)
		}

		count += *countOutput.Count
	}

	return count
}

func updateTableMoviesWriteThroughput() {
	sess := session.Must(session.NewSession())
	db := dynamodb.New(sess)

	moviesTableName := "movies"
	newWriteThroughput := int64(5)

	inputDescribeTable := dynamodb.DescribeTableInput{TableName: &moviesTableName}
	moviesTableDescribe, err := db.DescribeTable(&inputDescribeTable)
	if err != nil {
		log.Fatal(err)
	}

	moviesTable := moviesTableDescribe.Table
	currentProvisionedThroughput := moviesTable.ProvisionedThroughput

	newProvisionedThroughput := dynamodb.ProvisionedThroughput{
		ReadCapacityUnits:  currentProvisionedThroughput.ReadCapacityUnits,
		WriteCapacityUnits: &newWriteThroughput,
	}

	updateInput := dynamodb.UpdateTableInput{
		TableName:             &moviesTableName,
		ProvisionedThroughput: &newProvisionedThroughput,
	}

	_, err = db.UpdateTable(&updateInput)
	if err != nil {
		log.Fatal(err)
	}
}
