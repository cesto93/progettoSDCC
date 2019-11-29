#!/bin/bash
MASTER_NAME=master
WORKER_NAMES=worker-*
ID1=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$MASTER_NAME --output text --query 'Reservations[*].Instances[*].InstanceId')
ID2=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$WORKER_NAMES --output text --query 'Reservations[*].Instances[*].InstanceId')

aws ec2 start-instances --instance-ids $ID1
aws ec2 start-instances --instance-ids $ID2

sleep 10
source status_ec2.sh
