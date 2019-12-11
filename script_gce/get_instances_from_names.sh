#!/bin/bash
#create prometheus configuration from instances names

INPUT_JSON_FILE="../configuration/instances_names.json"
OUTPUT_JSON_FILE="../configuration/generated/instances.json"
mapfile -t INSTANCES <<< $(jq -r '.[]' $INPUT_JSON_FILE)
#gcloud compute instances describe ${INSTANCES[0]} --zone "us-central1-a" --format "json" | jq -s .
echo $(for i in "${INSTANCES[@]}"
do
	gcloud compute instances describe $i --zone "us-central1-a" --format "json" \
	| jq -s ".[].networkInterfaces[0].networkIP"
	#echo "$i"
done) | jq -s '[{targets:[(.[] + ":9100")],labels:{job:"prometheus"}}]'>$OUTPUT_JSON_FILE

