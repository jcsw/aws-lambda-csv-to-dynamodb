# Create lambda with terraform  to import csv to dynamodb

Build aws lambda

```text
cd aws-lambda/extract_movies_from_s3/cmd && GOOS=linux go build -o main && zip deployment.zip main && cd -
&&
cd aws-lambda/import_movies_in_dynamodb/cmd && GOOS=linux go build -o main && zip deployment.zip main && cd -
&&
cd aws-lambda/verify_movies_in_dynamodb/cmd && GOOS=linux go build -o main && zip deployment.zip main && cd -
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
S:imdb
N:year
S:title
S:code
]
```