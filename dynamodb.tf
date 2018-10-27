resource "aws_dynamodb_table" "movies" {
  name = "movies"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "imdb"
  range_key      = "year"

  attribute {
    name = "imdb"
    type = "S"
  }

  attribute {
    name = "year"
    type = "N"
  }

  tags {
    Name        = "Movies table"
    Environment = "Dev"
    Project     = "csv-to-dynamodb"
  }
}

resource "aws_appautoscaling_target" "movies_read" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/${aws_dynamodb_table.movies.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_target" "movies_write" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/${aws_dynamodb_table.movies.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  service_namespace  = "dynamodb"
}

