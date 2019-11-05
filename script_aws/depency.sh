#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
NAMES=( "master" "worker-1" "worker-2" "worker-3" )


for (( i=0; i<${#NAMES[@]}; i++ ));
do

INSTANCE_DNS=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].PublicDnsName')

echo "SSH CONNECTION TO $NAMES"



#connecting by ssh
ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "

#launching depency
sudo yum update -y
sudo yum install git -y
sudo yum install -y golang
go get -u github.com/aws/aws-sdk-go
"
#git clone https://github.com/cesto93/progettoSDCC

done
