#!/bin/bash
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
wget https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem -O /home/ec2-user/LikeMinds-Swarm/config/rds-combined-ca-bundle.pem
go build cmd/server/likeminds-swarm.go
