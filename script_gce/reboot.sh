#!/bin/bash

ZONE="us-central1-a"
CONF=$(<../configuration/word_count.json)

#importing configuration
NAMES=( $(echo $CONF | jq -r '.gc[].name') )

#ssh conn
for (( i=0; i<${#NAMES[@]}; i++ ));
do
gcloud compute ssh --zone=$ZONE ${NAMES[$i]} --command "sudo reboot" &
done
