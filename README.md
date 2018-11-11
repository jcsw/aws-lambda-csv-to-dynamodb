# Create lambda with terraform  to import csv to dynamodb

Build aws lambda

```text
cd aws-lambda/import_movies_from_s3_to_dynamodb && GOOS=linux go build -o main && zip deployment.zip main && cd -
```

Build terraform

```text
cd terraform
terraform init
terraform plan
terraform apply
```

table format:

```js
name: movies
fields: [
N:batchID
S:batchDate
S:imdb
N:year
S:title
S:code
]
```