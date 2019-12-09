# ProgettoSDCC

## Local Dependency
In order to launch the script you need to install

* jq
* konsole
* awscli
* gcloud

## AWS setup
For setup the instance to monitoring and wordcount application you need to:

*

## GC setup
In order to setup an instance running on google compute engine, you need to:

* launch the script setup_environment.sh in the folder progettoSDCC/script_gce specifying the instance name as an argument (es.: ./setup_environment instance-name-1), this will:
	- connect to the specified instance, 
	- install the necessary tools,
	- open an ssh connection to it
* for each instance you want to use as a monitor, after the ssh connection is established, launch the script set_instances_to_monitor.sh in the same folder of the previous point, specifying as arguments the names of the instance you want to monitor, including the monitor instance itself (es.: ./set_instances_to_monitor.sh instance-name-1 instance-name-2 ...), this will create some json files in the configuration folder:
	- a file containing the specified list (instances_names.json),
	- a file to let google know the IDs of instances to monitor (instances_ids.json),
	- a file to let prometheus know the IP's of instances to monitor (instances.json)

N.B.: 
-	by default only a subset of metrics are scraped by prometheus, gce and stackdriver; those metrics are specified in metrics_prometheus.json and metrics_gce.json, and can be expanded at will
-	only way for prometheus to connect to the monitored instances is via IP address, connecting through node exporter at port 9100, so ensure to start all instances you want to monitor and to allow both inbound and outbound traffic on the aforementioned port before running the script set_instances_to_monitor.sh (gce IPs are dynamically changed at each restart)

## Application

###Commands

The client of the application use different flags for specifing different operations or differrent args

* This are the commands for operation specification
    * -load this command load the files specified in the AWS S3 bucket at specified names
    * -delete this command delete the files specified by names on the AWS S3 bucket
    * -list this command list the files on the bucket
    * -count this command do the wordcount of the files in the bucket identified by given names

* This are the commands for arg specification:
    * -bucket (OPTIONAL) specified to use a bucket that MUST be existent
    * -names (USED with load/delete/count)specifies the names for the S3 file in the bucket to use/load/delete
    * -paths (USED with load) specifies the paths for the local file to upload
    * -serverAddr (USED with count) specifie the address of the server for the rpc requiest to wordcount operation

###Configuration
The node configuration is done via a json file located at /configuration/word_count.json  
The ip of the nodes are gathered by the script files and parse in a json file at /configuration/generated/app_node.json.  
To test the app in local or if you want to manually set the ip use this file directly.  

##Monitoring
