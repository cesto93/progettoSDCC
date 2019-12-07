#!/bin/bash
#get gce instances ids from names

INPUT_JSON_FILE="../configuration/instances_names.json"
OUTPUT_JSON_FILE="../configuration/generated/instances_ids.json"
mapfile -t INSTANCES <<< $(jq -r '.[]' $INPUT_JSON_FILE)
echo $(for i in "${INSTANCES[@]}"
do
	gcloud compute instances describe $i --zone "us-central1-a" --format "json" \
	| jq -s ".[].id"
	#echo "$i"
done) | jq -s '[.[]]'>$OUTPUT_JSON_FILE

