#!/bin/bash
MASTER_NAME=master
WORKER_NAMES=worker-*
echo MASTER
aws ec2 describe-instances --filters Name=tag:Name,Values=$MASTER_NAME --output text --query 'Reservations[*].Instances[*].State'
echo WORKERS
aws ec2 describe-instances --filters Name=tag:Name,Values=$WORKER_NAMES --output text --query 'Reservations[*].Instances[*].State'

