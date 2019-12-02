#!/bin/bash
NAME=$1
ID1=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$NAME --output text --query 'Reservations[*].Instances[*].InstanceId')
aws ec2 stop-instances --instance-ids $ID1
