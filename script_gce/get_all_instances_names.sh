#!/bin/bash
#create json with all reachable GCE instances names

JSON_FILE="../configuration/instances_names.json"

gcloud compute instances list --format "json" | jq -s 'map(.[].name)'>$JSON_FILE
