resource "aws_iam_role" "import_movies_from_s3_to_dynamodb_role" {
  name = "import_movies_from_s3_to_dynamodb_role"

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

resource "aws_iam_policy" "import_movies_from_s3_to_dynamodb_policy" {
    name        = "import_movies_from_s3_to_dynamodb_policy"
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
      "Resource": "arn:aws:s3::*:movies-csv/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:DescribeTable",
        "dynamodb:UpdateTable",
        "dynamodb:PutItem"
      ],
      "Resource": ["arn:aws:dynamodb:us-east-1:*:table/movies"]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "import_movies_from_s3_to_dynamodb_attach" {
    role       = "${aws_iam_role.import_movies_from_s3_to_dynamodb_role.name}"
    policy_arn = "${aws_iam_policy.import_movies_from_s3_to_dynamodb_policy.arn}"
}