#!/bin/bash
MASTER_NAME=master
WORKER_NAMES=worker-*
ID[0]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$MASTER_NAME  --query 'Reservations[*].Instances[*].InstanceId')
ID[1]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$WORKER_NAMES --query 'Reservations[*].Instances[*].InstanceId')

echo ${ID[@]} | jq -s 'add | flatten' >EC2_inst.json
