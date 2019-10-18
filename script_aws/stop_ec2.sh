#!/bin/bash
ID=$(aws ec2 describe-instances --filters 'Name=tag:Name,Values=monitor*' --output text --query 'Reservations[*].Instances[*].InstanceId')
aws ec2 stop-instances --instance-ids $ID 
