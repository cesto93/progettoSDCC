#!/bin/bash
#create json with GCE instances IPs and node_exporter port of all reachable instances
#use get_instances_complete if you need to get informations of a subset of instances (specified in file instances_names.json)

JSON_FILE="../configuration/generated/instances.json"		#same name declared in prometheus config

gcloud compute instances list --filter="status=running" --format "json" | jq -s 'map({targets:[(.[].networkInterfaces[0].accessConfigs[0].natIP + ":9100")],labels:{job:"prometheus"}})'>$JSON_FILE
