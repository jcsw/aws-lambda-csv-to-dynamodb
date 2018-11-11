package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	awsEvents "github.com/aws/aws-lambda-go/events"
	awsLambda "github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

type task struct {
	BatchID   string
	BatchDate string
	Record    []string
}

var moviesTableName = "movies"
var writeThroughputBeforeImport = int64(600)
var writeThroughputAfterImport = int64(5)
var timeToWaitTableRefreshThroughput = time.Duration(5) * time.Second

const numberOfWorkers = 4

var sess *session.Session

func main() {
	awsLambda.Start(handler)
}

func handler(ctx context.Context, s3Event awsEvents.S3Event) {
	fmt.Println("handler > s3Event:", s3Event)

	sess = session.Must(session.NewSession())

	fileObject, fileName := extractS3Object(s3Event)
	defer fileObject.Close()

	updateTableMoviesWriteThroughput(writeThroughputBeforeImport)

	processFile(fileObject, fileName)

	updateTableMoviesWriteThroughput(writeThroughputAfterImport)
}

func extractS3Object(s3Event awsEvents.S3Event) (io.ReadCloser, string) {

	s3Session := s3.New(sess)
	s3Entity := s3Event.Records[0].S3

	s3Object, err := s3Session.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3Entity.Bucket.Name),
		Key:    aws.String(s3Entity.Object.Key),
	})

	if err != nil {
		log.Fatal(err)
	}

	return s3Object.Body, s3Entity.Object.Key
}

func updateTableMoviesWriteThroughput(writeThroughputToImport int64) {

	dbSession := dynamodb.New(sess)
	moviesTableDescribe, err := dbSession.DescribeTable(&dynamodb.DescribeTableInput{TableName: &moviesTableName})
	if err != nil {
		log.Fatal(err)
	}

	moviesTable := moviesTableDescribe.Table
	currentProvisionedThroughput := moviesTable.ProvisionedThroughput

	newProvisionedThroughput := dynamodb.ProvisionedThroughput{
		ReadCapacityUnits:  currentProvisionedThroughput.ReadCapacityUnits,
		WriteCapacityUnits: &writeThroughputToImport,
	}

	updateInput := dynamodb.UpdateTableInput{
		TableName:             &moviesTableName,
		ProvisionedThroughput: &newProvisionedThroughput,
	}

	_, err = dbSession.UpdateTable(&updateInput)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(timeToWaitTableRefreshThroughput)
}

func processFile(fileObject io.ReadCloser, fileName string) {
	totalItems := 0

	batchID, batchDate := extractFieldsInFileName(fileName)
	fmt.Println("processFile > batchID:", batchID, "batchDate:", batchDate)

	reader := csv.NewReader(fileObject)
	header, _ := reader.Read()
	fmt.Println("processFile > csvHeader:", header)

	var dbSessions [numberOfWorkers]*dynamodb.DynamoDB
	for i := 0; i < numberOfWorkers; i++ {
		dbSessions[i] = dynamodb.New(sess)
	}

	tasks := make(chan task, numberOfWorkers)
	done := make(chan bool)

	for i := 0; i < numberOfWorkers; i++ {
		go func(worker int) {
			for {
				task, more := <-tasks
				if more {
					err := task.putItem(dbSessions[worker])
					if err != nil {
						fmt.Println("processFile > err:", err)
					}
				} else {
					fmt.Println("received all tasks, worker:", worker)
					done <- true
					return
				}
			}
		}(i)
	}

	printProcessStatus := true
	go func() {
		for printProcessStatus {
			fmt.Println("processFile > status batchID:", batchID, "batchDate:", batchDate, "totalItems:", totalItems)
			time.Sleep(time.Duration(1) * time.Minute)
		}
	}()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		totalItems++
		tasks <- task{
			BatchID:   batchID,
			BatchDate: batchDate,
			Record:    record,
		}
	}
	close(tasks)
	<-done
	printProcessStatus = false

	fmt.Println("processFile > finished batchID:", batchID, "batchDate:", batchDate, "totalItems:", totalItems)
}

func extractFieldsInFileName(fileName string) (string, string) {

	fileNameWithoutExtension := strings.Split(fileName, ".")[0]
	fileNameValues := strings.Split(fileNameWithoutExtension, "_")

	return fileNameValues[0], fileNameValues[1]
}

func (t *task) putItem(dbSession *dynamodb.DynamoDB) error {

	movieItem := makeMovieItemWithCSVRecord(t.BatchID, t.BatchDate, t.Record)

	_, err := dbSession.PutItem(&movieItem)

	if err != nil {
		return err
	}

	return nil
}

func makeMovieItemWithCSVRecord(batchID string, batchDate string, record []string) dynamodb.PutItemInput {

	return dynamodb.PutItemInput{
		TableName: aws.String(moviesTableName),
		Item: map[string]*dynamodb.AttributeValue{
			"imdb": {
				S: aws.String(record[0]),
			},
			"year": {
				N: aws.String(record[1]),
			},
			"title": {
				S: aws.String(record[2]),
			},
			"code": {
				S: aws.String(record[3]),
			},
			"batchID": {
				N: aws.String(batchID),
			},
			"batchDate": {
				S: aws.String(batchDate),
			},
		},
	}
}
