#!/usr/bin/env sh
set -eu

awslocal s3api create-bucket --bucket roliq-resumes-local
awslocal s3api put-bucket-cors --bucket roliq-resumes-local --cors-configuration '{"CORSRules":[{"AllowedOrigins":["http://localhost:3000"],"AllowedMethods":["PUT","HEAD"],"AllowedHeaders":["content-type","x-amz-meta-sha256","x-amz-*"],"ExposeHeaders":["ETag"],"MaxAgeSeconds":3600}]}'
awslocal sqs create-queue --queue-name roliq-domain-events
awslocal sns create-topic --name roliq-domain-events
