#!/bin/bash
source ./conf/key.sh
NAME=$1
INSTANCE_DNS=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$NAME --output text --query 'Reservations[*].Instances[*].PublicDnsName')
ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS
