package main

import (
 	"time"
 	"fmt"
 	"flag"
 	"progettoSDCC/source/monitoring/monitor"
 	"progettoSDCC/source/monitoring/zookeeper"
 	"progettoSDCC/source/monitoring/restarter"
 	"progettoSDCC/source/utility"
 )

 const (
 	zkServersIpPath = "../../configuration/zk_servers_addrs.json"
 	ec2InstPath = "../../configuration/ec2_inst.json"
 	membershipNodePath = "/membership"
 	sessionTimeout = 10
 	monitorInterval = 300
 )

 func saveMetrics(monitorBridge monitor.MonitorBridge) {
 	start := time.Now()
 	end := start.Add(time.Second * monitorInterval) 
 	ec2Data, err := monitorBridge.GetMetrics(start, end)
 	utility.CheckError(err)
 	printMetrics(ec2Data)
 }

 func checkMembersDead(zkBridge *zookeeper.ZookeeperBridge) {
 	for {
 		zkBridge.CheckMembersDead()
 	}
 }

 func printMetrics(results []monitor.MetricData) {
	for _, metricdata := range results {
		fmt.Println(metricdata.Label)
		for j, _ := range metricdata.Timestamps {
			fmt.Printf("%v %v\n", (metricdata.Timestamps[j]).String(), metricdata.Values[j])
		}
	} 
}

func main() {
	var zkServerAddresses, members []string
	var monitorBridge monitor.MonitorBridge
	var myRestarter restarter.Restarter
	var aws bool

	flag.BoolVar(&aws, "aws", false, "Specify the aws monitor")

 	/*startTime, _ := time.Parse(time.RFC3339, "2019-11-09T15:35:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T16:00:00+02:00")*/
 	startTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:00:00+00:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:10:00+00:00")

 	if (aws) {
 		monitorBridge = monitor.NewAws()
 		myRestarter = restarter.NewAws()
 	} else {
 		//TODO insert google monitor here
 		return
 	}

 	//load zk conf
 	utility.ImportJson(ec2InstPath, members)
 	utility.ImportJson(zkServersIpPath, zkServerAddresses)
 	zkServerAddresses = zkServerAddresses[0:3] //pick only 3 members for servers

 	zkBridge, err := zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, membershipNodePath, members)
 	utility.CheckError(err)
 	go checkMembersDead(zkBridge)

	ec2Data, err := monitorBridge.GetMetrics(startTime, endTime)
	utility.CheckError(err)
	printMetrics(ec2Data)

	for {
		if zkBridge.MembersDead != nil {
			for _, dead := range zkBridge.MembersDead {
				myRestarter.Restart(dead)
			}
		}
		saveMetrics(monitorBridge)
	}
 }