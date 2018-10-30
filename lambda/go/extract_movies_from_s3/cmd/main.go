package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type dbItem map[string]interface{}

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
	reader := csv.NewReader(result.Body)

	chunkSizeMax := 100
	chunkSize := 10

	items := make([]dbItem, 0, 0)

	header, _ := reader.Read()
	fmt.Println("header:", header)

	for {

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		fmt.Println("record:", record)

		item := makeItemByRecord(record)
		items = append(items, item)

		if len(items) == chunkSize {
			sendItems(items)
			items = make([]dbItem, 0, 0)

			if chunkSize < chunkSizeMax {
				chunkSize += 10
			}
		}
	}
}

func makeItemByRecord(record []string) map[string]interface{} {
	item := make(map[string]interface{})
	item["imdb"] = record[0]
	item["year"] = record[1]
	item["title"] = record[2]
	item["code"] = record[3]

	return item
}

func sendItems(items []dbItem) {
	fmt.Println("len:", len(items))
	fmt.Println("items:", items)
}
