resource "aws_dynamodb_table" "movies" {
  name = "movies"
  read_capacity  = 10
  write_capacity = 10
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
  max_capacity       = 20
  min_capacity       = 10
  resource_id        = "table/${aws_dynamodb_table.movies.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "movies_read_policy" {
  name               = "DynamoDBReadCapacityUtilization:${aws_appautoscaling_target.movies_read.resource_id}"
  policy_type        = "TargetTrackingScaling"
  resource_id        = "${aws_appautoscaling_target.movies_read.resource_id}"
  scalable_dimension = "${aws_appautoscaling_target.movies_read.scalable_dimension}"
  service_namespace  = "${aws_appautoscaling_target.movies_read.service_namespace}"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }

    target_value = 70
  }
}

resource "aws_appautoscaling_target" "movies_write" {
  max_capacity       = 220
  min_capacity       = 10
  resource_id        = "table/${aws_dynamodb_table.movies.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "movies_write_policy" {
  name               = "DynamoDBWriteCapacityUtilizationye:${aws_appautoscaling_target.movies_write.resource_id}"
  policy_type        = "TargetTrackingScaling"
  resource_id        = "${aws_appautoscaling_target.movies_write.resource_id}"
  scalable_dimension = "${aws_appautoscaling_target.movies_write.scalable_dimension}"
  service_namespace  = "${aws_appautoscaling_target.movies_write.service_namespace}"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }

    target_value = 30
  }
}
