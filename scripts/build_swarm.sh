#!/bin/bash
source /home/ec2-user/.profile
source /etc/environment

APPLICATION_AWS_S3_BUCKET=$PROJECT_AWS_S3_BUCKET
APPLICATION_DOT_ENV_KEY=$PROJECT_AWS_S3_DOT_ENV_KEY
APPLICATION_DOT_ENV_LOCATION=$PROJECT_DOT_ENV_LOCATION

cd /home/ec2-user/LikeMinds-Swarm/
branch_name="$(git rev-parse --abbrev-ref HEAD)"
if [ "$branch_name" = "development" ]
then
  git pull origin development
else
  git pull origin master
fi
aws s3api get-object --bucket "$APPLICATION_AWS_S3_BUCKET" --key "$APPLICATION_DOT_ENV_KEY" "$APPLICATION_DOT_ENV_LOCATION"
go build cmd/server/likeminds-swarm.go
