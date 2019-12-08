#!/bin/bash
#configure monitor datas
#arguments: list of instances to monitor (names)

X=($@)
printf '%s\n' "${X[@]}" | jq -R . | jq -s . > ../configuration/instances_names.json
source get_instances_complete.sh
