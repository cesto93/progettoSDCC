#!/bin/bash
KEY_POS="/home/pier/Desktop/progetto sdcc/myKey.pem"
NAME=$monitor*
INSTANCE_DNS=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$NAME --output text --query 'Reservations[*].Instances[*].PublicDnsName')
ssh -i "$KEY_POS" ec2-user@$INSTANCE_DNS
