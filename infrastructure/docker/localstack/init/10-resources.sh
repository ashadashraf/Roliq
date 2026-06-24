#!/usr/bin/env sh
set -eu

if ! awslocal s3api head-bucket --bucket roliq-resumes-local 2>/dev/null; then
  awslocal s3api create-bucket --bucket roliq-resumes-local
fi

awslocal s3api put-bucket-cors --bucket roliq-resumes-local --cors-configuration '{"CORSRules":[{"AllowedOrigins":["http://localhost:3000"],"AllowedMethods":["PUT","HEAD"],"AllowedHeaders":["*"],"ExposeHeaders":["ETag"],"MaxAgeSeconds":3600}]}'
awslocal sqs create-queue --queue-name roliq-domain-events
awslocal sns create-topic --name roliq-domain-events
