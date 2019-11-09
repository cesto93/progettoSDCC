#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
NAMES=( "master" "worker-1" "worker-2" "worker-3" )
GIT_KEY_FILE="sdcc_git"
LOCAL_DIR="/home/pier/.ssh"
AWS_DIR="/home/ec2-user/.ssh"

for (( i=0; i<${#NAMES[@]}; i++ ));
do

INSTANCE_DNS=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].PublicDnsName')

echo -e "COPYING SSH KEY FOR GIT DEPLOYMENT\n"

scp -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/$GIT_KEY_FILE  ec2-user@$INSTANCE_DNS:$AWS_DIR
scp -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/config  ec2-user@$INSTANCE_DNS:$AWS_DIR

echo -e "SSH CONNECTION TO $NAMES\n"

#connecting by ssh
ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "

sudo yum update -y
sudo yum install git -y
sudo yum install -y golang
go get -u github.com/aws/aws-sdk-go

#ssh -o 'StrictHostKeyChecking=no' -T git@github.com
git clone git@github.com:cesto93/progettoSDCC
"

done
