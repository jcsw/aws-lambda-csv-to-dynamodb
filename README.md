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