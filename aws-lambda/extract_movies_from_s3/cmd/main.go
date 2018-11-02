package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	lambdaService "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
)

type dbItem map[string]interface{}

type createMoviesEvent struct {
	SourceName string   `json:"sourceName"`
	Items      []dbItem `json:"items"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, s3Event events.S3Event) {

	record := s3Event.Records[0]
	s3Entity := s3Event.Records[0].S3

	fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3Entity.Bucket.Name, s3Entity.Object.Key)

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3Entity.Bucket.Name),
		Key:    aws.String(s3Entity.Object.Key),
	})

	if err != nil {
		log.Fatal(err)
	}

	defer result.Body.Close()
	processFile(result.Body, s3Entity.Object.Key)
}

func processFile(fileReader io.ReadCloser, fileName string) {

	reader := csv.NewReader(fileReader)

	chunkSizeMax := 100
	chunkSize := 10

	sleepTime := 1000
	sleepTimeMin := 500

	totalItems := 0
	totalchunks := 0

	items := make([]dbItem, 0, 0)

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
			sendItems(items, fileName)
			totalchunks++

			items = make([]dbItem, 0, 0)

			if chunkSize < chunkSizeMax {
				chunkSize += 10
			}

			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			if sleepTime > sleepTimeMin {
				sleepTime -= 100
			}
		}
	}

	if len(items) > 0 {
		sendItems(items, fileName)
		totalchunks++
	}

	fmt.Println("totalItems:", totalItems, "totalchunks:", totalchunks)
}

func makeItemByRecord(record []string) map[string]interface{} {
	item := make(map[string]interface{})
	item["imdb"] = record[0]
	item["year"] = record[1]
	item["title"] = record[2]
	item["code"] = record[3]

	return item
}

func sendItems(items []dbItem, fileName string) {
	fmt.Println("sendItems > len:", len(items))
	fmt.Println("sendItems > items:", items)

	event := createMoviesEvent{
		SourceName: fileName,
		Items:      items,
	}

	payload, err := json.Marshal(event)

	svc := lambdaService.New(session.New())
	input := &lambdaService.InvokeInput{
		FunctionName:   aws.String("import_movies_in_dynamodb"),
		InvocationType: aws.String("Event"),
		LogType:        aws.String("Tail"),
		Payload:        payload,
	}

	result, err := svc.Invoke(input)
	if err != nil {
		fmt.Println("sendItems > err:", err)
	}

	fmt.Printf("sendItems > statusCode:%d\n", result.StatusCode)
}
