resource "aws_lambda_function" "extract_movies_from_s3" {
  filename      = "../aws-lambda/extract_movies_from_s3/cmd/deployment.zip"
  function_name = "extract_movies_from_s3"
  role          = "${aws_iam_role.extract_movies_from_s3_role.arn}"
  handler       = "main"
  runtime       = "go1.x"
  timeout       = "300"
  reserved_concurrent_executions = 1
}

resource "aws_lambda_function" "import_movies_in_dynamodb" {
  filename      = "../aws-lambda/import_movies_in_dynamodb/cmd/deployment.zip"
  function_name = "import_movies_in_dynamodb"
  role          = "${aws_iam_role.import_movies_in_dynamodb_role.arn}"
  handler       = "main"
  runtime       = "go1.x"
  timeout       = "120"
  reserved_concurrent_executions = 6
}