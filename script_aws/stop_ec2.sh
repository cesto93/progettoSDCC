#!/bin/bash
MASTER_NAME=master
WORKER_NAMES=worker-*
ID1=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$MASTER_NAME --output text --query 'Reservations[*].Instances[*].InstanceId')
ID2=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$WORKER_NAMES --output text --query 'Reservations[*].Instances[*].InstanceId')

aws ec2 stop-instances --instance-ids $ID1 $ID2
