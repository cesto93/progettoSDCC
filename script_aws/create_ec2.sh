#!/bin/bash
#parameter start
IMG=ami-0b69ea66ff7391e80
TYPE=t2.micro
KEY=myKey
GROUP=my_group
GROUP_ID=sg-04d681a46dfb3274e
INSTANCE_NAMES=( "master" "worker-1" "worker-2" "worker-3" )
#parameter ends
#creating instances
for (( i=0; i<${#INSTANCE_NAMES[@]}; i++ ));
do
EC2_RESP=$(aws ec2 run-instances --image-id $IMG --count 1 --instance-type $TYPE --key-name $KEY --security-group-ids $GROUP_ID \
		--tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=${INSTANCE_NAMES[$i]}}]")
done



