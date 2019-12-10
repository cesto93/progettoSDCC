#!/bin/bash
#set up environment
#argument: instance name

ZONE="us-central1-a"
GROUP="instance-group-1"
CONF_MONITOR=$(<../configuration/monitor.json)
MONITOR_NAMES=( $(echo $CONF_MONITOR | jq -r '.gc[].name') )

source ./conf/key.sh

# Generating Prometheus configuration file
tee ../configuration/generated/prometheus.yml<<EOF 1> /dev/null
# my global config
global:
  scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
  # scrape_timeout is set to the global default (10s).

scrape_configs:
  - job_name: 'MonitorInstances'
    file_sd_configs:
      - files:
        - $HOME_DIR/go/src/progettoSDCC/configuration/generated/instances.json
EOF
for k in ${MONITOR_NAMES[@]};
do
	echo $k
	konsole --new-tab --noclose -e ./setup_single_monitor.sh $k &
done;
