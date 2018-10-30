# Create lambda with terraform  to import csv to dynamodb

Create infra
```
terraform init
terraform plan
terraform apply
```

Upload movies_1.csv in bucket ```import.movies.csv```

table: ```movies```
```
S:imdb
N:year
S:title
S:code
```

Build go lambda
```
cd lambda/go/extract_movies_from_s3/cmd
GOOS=linux go build -o main
zip deployment.zip main
```