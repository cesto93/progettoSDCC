#!/bin/bash
#create json with GCE instances name

JSON_FILE="../configuration/instances_names.json"

gcloud compute instances list --format "json" | jq -s 'map(.[].name)'>$JSON_FILE
