resource "aws_iam_role" "extract_movies_from_s3_role" {
  name = "extract_movies_from_s3_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {"Service": "lambda.amazonaws.com"}
    }
  ]
}
EOF
}

resource "aws_iam_policy" "extract_movies_from_s3_policy" {
    name        = "extract_movies_from_s3_policy"
    description = "Lambda to extract movies policy"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:us-east-1:*:*"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": ["lambda:InvokeFunction"],
      "Resource": ["*"]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "extract_movies_from_s3_attach" {
    role       = "${aws_iam_role.extract_movies_from_s3_role.name}"
    policy_arn = "${aws_iam_policy.extract_movies_from_s3_policy.arn}"
}

resource "aws_lambda_function" "extract_movies_from_s3" {
  filename      = "lambda/extract_movies_from_s3/function.zip"
  function_name = "extract_movies_from_s3"
  role          = "${aws_iam_role.extract_movies_from_s3_role.arn}"
  handler       = "extract_movies_from_s3.handler"
  runtime       = "nodejs8.10"
  timeout       = "60"
}

resource "aws_iam_role" "import_movies_in_dynamodb_role" {
  name = "import_movies_in_dynamodb_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {"Service": "lambda.amazonaws.com"},
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "import_movies_in_dynamodb_policy" {
    name        = "import_movies_in_dynamodb_policy"
    description = "Lambda to import movies in dynamodb"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:us-east-1:*:*"
    },
    {
      "Effect": "Allow",
      "Action": ["dynamodb:PutItem"],
      "Resource": ["arn:aws:dynamodb:us-east-1:*:table/movies"]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "import_movies_in_dynamodb_attach" {
    role       = "${aws_iam_role.import_movies_in_dynamodb_role.name}"
    policy_arn = "${aws_iam_policy.import_movies_in_dynamodb_policy.arn}"
}

resource "aws_lambda_function" "import_movies_in_dynamodb" {
  filename      = "lambda/import_movies_in_dynamodb/function.zip"
  function_name = "import_movies_in_dynamodb"
  role          = "${aws_iam_role.import_movies_in_dynamodb_role.arn}"
  handler       = "import_movies_in_dynamodb.handler"
  runtime       = "nodejs8.10"
  timeout       = "60"
  reserved_concurrent_executions = 10
}