#!/bin/bash
#create json with GCE instances IPs and node_exporter port

gcloud compute instances list --filter="status=running" --format "json" | jq -s 'map({targets:[(.[].networkInterfaces[0].accessConfigs[0].natIP + ":9100")],labels:{job:"prometheus"}})'>instances.json