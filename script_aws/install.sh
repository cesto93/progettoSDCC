#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
NAMES=( "master" "worker-1" "worker-2" "worker-3" )

for (( i=0; i<${#NAMES[@]}; i++ ));
do

INSTANCE_DNS=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].PublicDnsName')

echo "SSH CONNECTION TO ${NAMES[$i]} AT $INSTANCE_DNS"

#connecting by ssh
ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "

echo -e DOWNLOADING SOURCE\n

cd ./go/src/progettoSDCC
#ssh -o 'StrictHostKeyChecking=no' -T git@github.com
git pull git@github.com:cesto93/progettoSDCC

go build -o ./source/application/word_counter/worker/worker ./source/application/word_counter/worker/worker.go
go build -o ./source/application/word_counter/master/master ./source/application/word_counter/master/master.go
"
done

