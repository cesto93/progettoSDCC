package main

import (
 	"time"
 	"fmt"
 	"flag"
 	"log"
 	"progettoSDCC/source/monitoring/monitor"
 	"progettoSDCC/source/monitoring/zookeeper"
 	"progettoSDCC/source/monitoring/restarter"
 	"progettoSDCC/source/utility"
 )

 const (
 	sessionTimeout = 10
 	monitorIntervalSeconds = 300
 	zkServersIpPath = "../configuration/generated/zk_servers_addrs.json"
 	zkAgentPath = "../configuration/generated/zk_agent.json"
 	aliveNodePath = "/alive"
 	startingNodePath = "/starting"

 	EC2MetricJsonPath = "../configuration/metrics_ec2.json"
 	EC2InstPath = "../configuration/generated/ec2_inst.json"
 	S3MetricPath = "../configuration/metrics_s3.json"
 	StatPath = "../configuration/monitoring_stat.json"
 )

 func saveMetrics(monitorBridge monitor.MonitorBridge, interval time.Duration) {
 	end := time.Now()
 	start := end.Add(-interval) 
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
 	fmt.Println("Metrics:\n")
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
	var index int

	flag.BoolVar(&aws, "aws", false, "Specify the aws monitor")
	flag.IntVar(&index, "index", 0, "Specify the index of the agent")
	flag.Parse()

 	/*startTime, _ := time.Parse(time.RFC3339, "2019-11-09T15:35:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T16:00:00+02:00")
 	startTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:00:00+00:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:10:00+00:00")*/

 	if (aws) {
 		monitorBridge = monitor.NewAws(EC2MetricJsonPath, EC2InstPath, S3MetricPath, StatPath)
 		myRestarter = restarter.NewAws()
 	} else {
 		//TODO insert google monitor here
 		log.Fatal("google agent not implemented yes\n")
 	}

 	//load zk conf
 	err := utility.ImportJson(zkAgentPath, &members)
 	utility.CheckError(err)
 	err = utility.ImportJson(zkServersIpPath, &zkServerAddresses)
 	utility.CheckError(err)

 	fmt.Println(zkServerAddresses)
 	//less than 3 servers dosen't make zookeeper fault tolerant

 	zkBridge, err := zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, aliveNodePath, startingNodePath,
 									 members)
 	utility.CheckError(err)
 	err = zkBridge.RegisterMember(members[index], "info")
 	go checkMembersDead(zkBridge)
 	utility.CheckError(err)

	/*ec2Data, err := monitorBridge.GetMetrics(startTime, endTime)
	utility.CheckError(err)
	printMetrics(ec2Data)*/

	monitorInterval := monitorIntervalSeconds * time.Second
	now := time.Now()
	saveMetrics(monitorBridge, monitorInterval)
	nextMeasure := time.Now()
	for {
		time.Sleep(time.Second)
		if zkBridge.MembersDead != nil {
			for _, dead := range zkBridge.MembersDead {
				fmt.Println("This is is dead: " + dead + "\n")
				err = myRestarter.Restart(dead)
				fmt.Println(err)
				zkBridge.MemberIsStarting(dead, time.Now().String())
				//utility.CheckError(err)
			}
		}
		if (now.After(nextMeasure)) {
			saveMetrics(monitorBridge, monitorInterval)
		}
	}
 }