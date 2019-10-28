#!/bin/bash
NAME=$monitor*
ID=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$NAME --output text --query 'Reservations[*].Instances[*].InstanceId')
aws ec2 start-instances --instance-ids $ID
