resource "aws_lambda_function" "import_movies_from_s3_to_dynamodb" {
  filename      = "../aws-lambda/import_movies_from_s3_to_dynamodb/deployment.zip"
  function_name = "import_movies_from_s3_to_dynamodb"
  role          = "${aws_iam_role.import_movies_from_s3_to_dynamodb_role.arn}"
  handler       = "main"
  runtime       = "go1.x"
  timeout       = "900"
  memory_size   = "512"
  reserved_concurrent_executions = 1
}