#!/bin/bash
MASTER_NAME=master
WORKER_NAMES=worker-*
ID1=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$MASTER_NAME  --query 'Reservations[*].Instances[*].InstanceId')
ID2=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$WORKER_NAMES --query 'Reservations[*].Instances[*].InstanceId')

echo $ID1 $ID2 | jq -s 'add | flatten' >EC2_inst.json
