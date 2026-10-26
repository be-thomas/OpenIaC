resource "aws_iam_user" "default" {
  name = coalesce(var.bucket_iam_user_name, var.bucket_name)
}

resource "aws_iam_access_key" "default" {
  user = aws_iam_user.default.name
}

resource "aws_iam_policy" "bucket" {
  name = "s3-${var.bucket_name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "polly:*"
        ],
        "Resource": [
            "*"
        ]
    },
    {
        "Effect": "Allow",
        "Action": [
            "s3:ListBucket"
        ],
        "Resource": [
            "${module.bucket.s3_bucket_arn}"
        ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:PutObjectAcl",
        "s3:GetObjectAcl",
        "s3:DeleteObject",
        "s3:GetObject"
      ],
      "Resource": "${module.bucket.s3_bucket_arn}/*"
    }
  ]
}
EOF

}

resource "aws_iam_user_policy_attachment" "bucket" {
  user       = aws_iam_user.default.name
  policy_arn = aws_iam_policy.bucket.arn
}

