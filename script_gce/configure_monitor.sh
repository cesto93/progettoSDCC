#!/bin/bash

ZONE="us-central1-a"
CONF_MONITOR=$(<../configuration/monitor.json)
MONITOR_NAMES=( $(echo $CONF_MONITOR | jq -r '.gc[].name') )

for ((i=0; i<${#MONITOR_NAMES[@]}; i++));
do
	MONITORED_NAMES=( $(echo $CONF_MONITOR | jq -r --argjson index "$i" '.gc[$index].monitored[]') )
	M=$(echo ${MONITORED_NAMES[@]})
	gcloud compute ssh --zone=$ZONE ${MONITOR_NAMES[i]} \
	--command "cd ~/go/src/progettoSDCC/script_gce && export PATH=$PATH:/snap/bin && bash ./set_instances_to_monitor.sh $M"
done
