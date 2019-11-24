package main

import (
 	"time"
 	"log"
 	"progettoSDCC/source/monitoring/monitor"
 	"progettoSDCC/source/monitoring/zookeeper"
 	"progettoSDCC/source/utility"
 )

 const (
 	zkServersIpPath = "../../configuration/zk_servers_addrs.json"
 	monitorMembersPath = "../../configuration/monitor_members.json"
 	membershipNodePath = "/membership"
 	sessionTimeout = 10
 	monitorInterval = 300
 	aws = true
 )

 func saveMetrics(monitorBridge monitor.MonitorBridge) {
 	start := time.Now()
 	end := start.Add(time.Second * monitorInterval) 
 	ec2Data := monitorBridge.GetMetrics(start, end)
 	monitor.PrintMetrics(ec2Data)
 }

 func checkMembersAlive(zkBridge *zookeeper.ZookeeperBridge) {
 	for {
 		zkBridge.CheckMembers()
 	}
 }

func main() {
	var zkServerAddresses, members []string
	var monitorBridge monitor.MonitorBridge
 	/*startTime, _ := time.Parse(time.RFC3339, "2019-11-09T15:35:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T16:00:00+02:00")*/
 	startTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:00:00+00:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:10:00+00:00")

 	utility.ImportJson(zkServersIpPath, zkServerAddresses)
 	utility.ImportJson(monitorMembersPath, members)
 	
 	if (aws) {
 		monitorBridge = monitor.New()
 	} else {
 		//TODO insert google monitor here
 		return
 	}

 	zkBridge, err := zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, membershipNodePath, members)

 	if (err != nil) {
 		log.Fatal("Error in zkBridge generation: ", err)
 	}
 	go checkMembersAlive(zkBridge)

	ec2Data := monitorBridge.GetMetrics(startTime, endTime)
	monitor.PrintMetrics(ec2Data)

 }