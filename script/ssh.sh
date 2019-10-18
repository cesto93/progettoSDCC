#!/bin/bash
KEY_POS="./myKey.pem"
INSTANCES=$(aws ec2 describe-instances)
INSTANCE_DNS=$(echo $INSTANCES | jq -r '.Reservations[].Instances[].PublicDnsName')
ssh -i $KEY_POS ec2-user@$INSTANCE_DNS
