package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
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
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
)

type dbItem map[string]interface{}

type importMoviesEvent struct {
	SourceName string   `json:"sourceName"`
	BatchID    string   `json:"batchID"`
	BatchDate  string   `json:"batchDate"`
	Items      []dbItem `json:"items"`
}

type verifyMoviesEvent struct {
	SourceName string `json:"sourceName"`
	BatchID    string `json:"batchID"`
	BatchDate  string `json:"batchDate"`
	TotalItems int    `json:"totalItems"`
}

func main() {
	awsLambda.Start(handler)
}

func handler(ctx context.Context, s3Event awsEvents.S3Event) {

	record := s3Event.Records[0]
	s3Entity := s3Event.Records[0].S3

	fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3Entity.Bucket.Name, s3Entity.Object.Key)

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	s3Object, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3Entity.Bucket.Name),
		Key:    aws.String(s3Entity.Object.Key),
	})

	if err != nil {
		log.Fatal(err)
	}
	defer s3Object.Body.Close()

	updateTableMoviesWriteThroughput()

	processFile(s3Object.Body, s3Entity.Object.Key)
}

func updateTableMoviesWriteThroughput() {
	sess := session.Must(session.NewSession())
	db := dynamodb.New(sess)

	moviesTableName := "movies"
	newWriteThroughput := int64(300)
	timeInSecondsWaitTableRefresh := 5

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

	output, err := db.UpdateTable(&updateInput)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("updateTableMoviesWriteThroughput > output:", output)

	time.Sleep(time.Duration(timeInSecondsWaitTableRefresh) * time.Second)
}

func processFile(fileReader io.ReadCloser, fileName string) {

	reader := csv.NewReader(fileReader)

	chunkSize := 300

	totalItems := 0
	totalchunks := 0

	items := make([]dbItem, 0, 0)

	fileNameWithoutExtension := strings.Split(fileName, ".")[0]
	fileNameValues := strings.Split(fileNameWithoutExtension, "_")

	batchID := fileNameValues[0]
	batchDate := fileNameValues[1]

	fmt.Println("processFile > init batchID:", batchID, "batchDate:", batchDate)

	header, _ := reader.Read()
	fmt.Println("processFile > header:", header)

	for {

		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		totalItems++
		item := makeItemByRecord(record)
		items = append(items, item)

		if len(items) == chunkSize {
			totalchunks++
			sendToImport(fileName, batchID, batchDate, items)
			items = make([]dbItem, 0, 0)
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
	}

	if len(items) > 0 {
		sendToImport(fileName, batchID, batchDate, items)
		totalchunks++
	}

	sendToVerify(fileName, batchID, batchDate, totalItems)

	fmt.Println("processFile > finished batchID:", batchID, "batchDate:", batchDate, "totalItems:", totalItems, "totalchunks:", totalchunks)
}

func makeItemByRecord(record []string) map[string]interface{} {
	item := make(map[string]interface{})
	item["imdb"] = record[0]
	item["year"] = record[1]
	item["title"] = record[2]
	item["code"] = record[3]

	return item
}

func sendToImport(fileName string, batchID string, batchDate string, items []dbItem) {
	event := importMoviesEvent{
		SourceName: fileName,
		BatchID:    batchID,
		BatchDate:  batchDate,
		Items:      items,
	}

	invokeEventFunction("import_movies_in_dynamodb", event)
}

func sendToVerify(fileName string, batchID string, batchDate string, totalItems int) {
	event := verifyMoviesEvent{
		SourceName: fileName,
		BatchID:    batchID,
		BatchDate:  batchDate,
		TotalItems: totalItems,
	}

	invokeEventFunction("verify_movies_in_dynamodb", event)
}

func invokeEventFunction(functionName string, event interface{}) {
	fmt.Println("invokeEventFunction > functionName:", functionName, "event:", event)

	payload, err := json.Marshal(event)

	if err != nil {
		log.Fatal(err)
	}

	svc := lambda.New(session.New())
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: aws.String("Event"),
		LogType:        aws.String("Tail"),
		Payload:        payload,
	}

	result, err := svc.Invoke(input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("invokeEventFunction > resultStatusCode:", *result.StatusCode)
}
