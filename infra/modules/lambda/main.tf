data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "history_bucket_access" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]

    resources = [
      "${var.history_bucket_arn}/*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      var.history_bucket_arn,
    ]
  }
}

resource "aws_iam_role" "lambda" {
  name               = "${var.function_name}-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "basic_execution" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_policy" "history_bucket_access" {
  name   = "${var.function_name}-history-bucket-access"
  policy = data.aws_iam_policy_document.history_bucket_access.json
}

resource "aws_iam_role_policy_attachment" "history_bucket_access" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.history_bucket_access.arn
}

resource "aws_lambda_function" "backend" {
  function_name = var.function_name
  role          = aws_iam_role.lambda.arn
  filename      = var.package_file

  architectures = ["x86_64"]
  handler       = "bootstrap"
  runtime       = "provided.al2023"

  memory_size = 128
  timeout     = 10

  source_code_hash = filebase64sha256(var.package_file)

  environment {
    variables = {
      HISTORY_BUCKET = var.history_bucket_name
    }
  }
}

resource "aws_lambda_function_url" "backend" {
  function_name      = aws_lambda_function.backend.function_name
  authorization_type = "NONE"

  cors {
    allow_methods = ["GET", "POST"]
    allow_origins = ["*"]
  }
}

resource "aws_lambda_permission" "function_url" {
  statement_id           = "AllowPublicFunctionUrlInvoke"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.backend.function_name
  principal              = "*"
  function_url_auth_type = "NONE"
}

resource "aws_lambda_permission" "function_url_invoke" {
  statement_id  = "AllowPublicFunctionInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.backend.function_name
  principal     = "*"
}
