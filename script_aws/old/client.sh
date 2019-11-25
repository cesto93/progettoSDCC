#!/bin/bash

MASTER=( "master" )
PORT=( "1050" )
files=( "prova1.txt" "prova2.txt" "gpl-3.0.txt" )

IP=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$MASTER --output text \
		--query 'Reservations[*].Instances[*].NetworkInterfaces[*].Association.PublicIp')
ADDRESS="$IP:$PORT"

cd ../source/application/word_counter

echo "./client/client -count -names ${files[0]},${files[1]},${files[2]} -serverAddr $ADDRESS"
./client/client -count -names ${files[0]},${files[1]},${files[2]} -serverAddr $ADDRESS
