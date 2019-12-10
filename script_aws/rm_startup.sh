#!/bin/bash
CONF_MONITOR=$(<../configuration/monitor.json)

#importing configuration
MONITOR_NAMES=( $(echo $CONF_MONITOR | jq -r '.aws[].name') )

base64 $SCRIPT > $SCRIPT_BASE64

for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${MONITOR_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	ID_MONITOR[$i]=$(echo $INST | jq -r '.[].InstanceId')
	aws ec2 modify-instance-attribute --instance-id ${ID_MONITOR[$i]} --user-data ":"
done


