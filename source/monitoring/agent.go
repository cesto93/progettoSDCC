package main

import (
 	"time"
 	"fmt"
 	"flag"
 	"progettoSDCC/source/monitoring/monitor"
 	"progettoSDCC/source/monitoring/zookeeper"
 	"progettoSDCC/source/monitoring/restarter"
 	"progettoSDCC/source/monitoring/db"
 	"progettoSDCC/source/utility"
 )

 const (
 	sessionTimeout = 10
 	monitorIntervalSeconds = 300
 	restartIntervalSecond = 5

 	dbAddrPath = "../configuration/generated/db_addr.json"
 	dbName = "mydb"

 	zkServersIpPath = "../configuration/generated/zk_servers_addrs.json"
 	zkAgentPath = "../configuration/generated/zk_agent.json"
 	idMonitorPath = "../configuration/generated/id_monitor.json"
 	aliveNodePath = "/alive"


 	EC2MetricJsonPath = "../configuration/metrics_ec2.json"
 	EC2InstPath = "../configuration/generated/ec2_inst.json"
 	S3MetricPath = "../configuration/metrics_s3.json"
 	StatPath = "../configuration/monitoring_stat.json"

 	GcloudMetricsJsonPath = "../../configuration/metrics_gce.json"
    InstancesJsonPath = "../../configuration/generated/instances_ids.json"

    PrometheusMetricsJsonPath = "../../configuration/metrics_prometheus.json"
 )

 func saveMetrics(monitorBridge monitor.MonitorBridge, dbBridge *db.DbBridge, interval time.Duration) {
 	end := time.Now()
 	start := end.Add(-interval) 
 	ec2Data, err := monitorBridge.GetMetrics(start, end)
 	utility.CheckError(err)
 	printMetrics(ec2Data)
 	dbBridge.SaveMetrics(ec2Data)
 }

 func checkMembersDead(zkBridge *zookeeper.ZookeeperBridge, id string) {
 	for {
 		zkBridge.CheckMemberDead(id)
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
	var dbAddr string
	var monitorBridge, monitorPrometheus monitor.MonitorBridge
	var myRestarter restarter.Restarter
	var aws,tryed bool
	var index, next int
	var now, nextMeasure time.Time

	flag.BoolVar(&aws, "aws", false, "Specify the aws monitor")
	flag.Parse()

	// db conf
	err := utility.ImportJson(dbAddrPath, &dbAddr)
	utility.CheckError(err)
	dbBridge:= db.NewDb(dbAddr, dbName)

	// zk conf
 	err = utility.ImportJson(zkAgentPath, &members)
 	utility.CheckError(err)
 	err = utility.ImportJson(zkServersIpPath, &zkServerAddresses)
 	utility.CheckError(err)
 	err = utility.ImportJson(idMonitorPath, &index)
 	utility.CheckError(err)
 	next = (index + 1) % len(members) //this is the id of agent to restart if crash
 	fmt.Println(zkServerAddresses) //less than 3 servers dosen't make zookeeper fault tolerant

 	zkBridge, err := zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, aliveNodePath, members)
 	utility.CheckError(err)
 	err = zkBridge.RegisterMember(members[index], "info")
 	utility.CheckError(err)
 	go checkMembersDead(zkBridge, members[next])

 	if (aws) {
 		monitorBridge = monitor.NewAws(EC2MetricJsonPath, EC2InstPath, S3MetricPath, StatPath)
 		myRestarter = restarter.NewAws()
 	} else {
 		monitorBridge = monitor.NewGce(GcloudMetricsJsonPath, InstancesJsonPath)
 		myRestarter = restarter.NewGce()
 	}
 	monitorPrometheus = monitor.NewPrometheus(PrometheusMetricsJsonPath)

 	//get last five minutes time range
	startTime := time.Now().UTC().Add(time.Minute * -5)
    endTime := time.Now().UTC()
 	monitorBridge.GetMetrics(startTime, endTime)
	monitorPrometheus.GetMetrics(startTime, endTime)

	monitorInterval := monitorIntervalSeconds * time.Second
	nextMeasure = now.Add(-monitorInterval)
	for {
		time.Sleep(time.Second * restartIntervalSecond)
		now = time.Now()
		if zkBridge.IsDead != false {
			fmt.Println("This is is dead: " + members[next])
			tryed, err = myRestarter.Restart(members[next])
			fmt.Println(err)
			//utility.CheckError(err)

			if tryed {
				fmt.Println("Tried to restart")
			}
		}
		if (now.After(nextMeasure)) {
			saveMetrics(monitorBridge, dbBridge, monitorInterval)
			nextMeasure = now.Add(monitorInterval)
		}
	}
 }