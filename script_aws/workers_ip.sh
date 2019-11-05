#!/bin/bash

NAMES=( "worker-1" "worker-2" "worker-3" )
PORTS=( "1050" "1051" "1052" )
files=( "prova1.txt" "prova2.txt" "gpl-3.0.txt" )
for (( i=0; i<${#NAMES[@]}; i++ ));
do

IP[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].PublicDnsName')
ADDRESSES=$"${IP[$i]}:${PORTS[$i]}"
done

cd ../../

./master/master -files ${files[0]},${files[1]},${files[2]} \
-ports ${addresses[0]},${addresses[1]},${addresses[2]}

