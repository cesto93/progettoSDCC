#!/bin/bash
#create prometheus configuration from instances names

INPUT_JSON_FILE="../configuration/instances_names.json"
OUTPUT_JSON_FILE="../configuration/instances.json"
mapfile -t INSTANCES <<< $(jq -r '.[]' $INPUT_JSON_FILE)
echo $(for i in "${INSTANCES[@]}"
do
	gcloud compute instances describe $i --zone "us-central1-a" --format "json" \
	| jq -s ".[].networkInterfaces[0].accessConfigs[0].natIP"
	#echo "$i"
done) | jq -s '[{targets:[(.[] + ":9100")],labels:{job:"prometheus"}}]'>$OUTPUT_JSON_FILE

