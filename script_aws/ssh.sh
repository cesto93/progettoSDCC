#!/bin/bash
KEY_POS="./myKey.pem"
INSTANCE_DNS=$(aws ec2 describe-instances --filters 'Name=tag:Name,Values=monitor*' --output text --query 'Reservations[*].Instances[*].PublicDnsName')
ssh -i $KEY_POS ec2-user@$INSTANCE_DNS
