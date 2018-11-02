resource "aws_s3_bucket" "movies_csv_import" {
  bucket = "movies.csv.import"
  acl    = "private"
  force_destroy = true

  tags {
    Name        = "Movies CSV bucket"
    Environment = "Dev"
    Project     = "csv-to-dynamodb"
  }
}

resource "aws_lambda_permission" "extract_movies_from_s3_allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.extract_movies_from_s3.arn}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${aws_s3_bucket.movies_csv_import.arn}"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = "${aws_s3_bucket.movies_csv_import.id}"

  lambda_function {
    lambda_function_arn = "${aws_lambda_function.extract_movies_from_s3.arn}"
    events              = ["s3:ObjectCreated:*"]
    filter_suffix       = ".csv"
  }
}