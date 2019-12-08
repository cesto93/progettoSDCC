package main

import (
	"fmt"
	"log"
	"net/rpc"
	"flag"
	"time"
	"progettoSDCC/source/application/word_counter/storage"
	"progettoSDCC/source/application/word_counter/rpcUtils"
	"progettoSDCC/source/application/word_counter/wordCountUtils"
	"progettoSDCC/source/utility"
	"progettoSDCC/source/metrics"
)

type Master int

var nodeConf rpcUtils.NodeConfiguration
var bucketName string

const (
	nodesJsonPath = "../configuration/generated/app_node.json"
	metricsJsonPath = "../log/app_metrics.json"
)

func readWordfilesFromS3(keys []string, bucketName string) []string {
	var texts []string
	s := storage.New(bucketName)
	for _,key := range keys {
		data,err := s.Read(key)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		texts = append(texts, string(data))
	}
	return texts
}

func saveResults(res []wordCountUtils.WordCount, name string){
	s := storage.New(bucketName)
	text := wordCountUtils.ToString(res)
	err := s.Write(name, []byte(text))
	if err != nil {
			log.Fatal(err)
	}
}

//ASYNC
func callMapOnWorkers(texts []string, nodes []rpcUtils.Node) error {
	rpcChan := make(chan *rpc.Call, len(nodes))
	for i := range texts {
		client, err := rpc.DialHTTP("tcp", nodes[i % len(nodes)].Address + ":" + nodes[i % len(nodes)].Port)
		if err != nil {
			return fmt.Errorf("Error in dialing: %v", err)
		}
		defer client.Close()
		client.Go("Worker.Map", texts[i], nil, rpcChan)
	}
	for i := range texts {
		divCall := <-rpcChan
		if divCall.Error != nil {
			return fmt.Errorf("Error in rpcMap worker %d : %v ", i % len(nodes), " :", divCall.Error.Error())
		}
	}
	return nil
}

//ASYNC
func callBarrierOnWorkers(nodes []rpcUtils.Node) error {
	rpcChan := make(chan *rpc.Call, len(nodes))
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			return fmt.Errorf("Error in dialing: %v", err)
		}
		defer client.Close()
		state := true
		client.Go("Worker.EndMapFase", state, nil, rpcChan)
	}
	for i := range nodes {
		divCall := <-rpcChan
		if divCall.Error != nil {
			return fmt.Errorf("Error in rpcEndMapFase num %d: %v", i, " :", divCall.Error.Error())
		}
	}
	return nil
}

//SYNC
func callLoadTopologyOnWorker(topology []rpcUtils.Node, node rpcUtils.Node) error {
	client, err := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
	if err != nil {
		return fmt.Errorf("Error in dialing: %v", err)
	}
	defer client.Close()
	err = client.Call("Worker.LoadTopology", topology, nil)
	if err != nil {
		return fmt.Errorf("Error in rpcMap: %v", err)
	}
	return nil
}

//SYNC
func getResultsOnWorkers(nodes []rpcUtils.Node) ([]wordCountUtils.WordCount, error) {
	var res []wordCountUtils.WordCount
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			return nil, fmt.Errorf("Error in dialing: %v", err)
		}
		defer client.Close()
		var counted []wordCountUtils.WordCount
		state := true
		err = client.Call("Worker.GetResults", state, &counted)
		if err != nil {
			return nil, fmt.Errorf("Error in rpcMap: %v", err)
		}
		//fmt.Println("words by reducer ", i, " = ", wordCountUtils.CountTotalWords(counted))

		for j := range counted {
			res = append(res, counted[j])
		}
	}
	return res, nil
}

//SYNC
func checkWorker(node rpcUtils.Node) error {
	client, err := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
	if err != nil {
		return fmt.Errorf("Error in dialing: %v", err)
	}
	defer client.Close()
	err = client.Call("Worker.CheckConn", true, nil)
	if err != nil {
		return fmt.Errorf("Error in rpcMap: %v", err)
	}
	return nil
}

func (t *Master) DoWordCount(wordFiles []string, res *bool) error{
	var err error
	fmt.Println("Request received")
	start := time.Now()

	nodes := nodeConf.Workers
	words := readWordfilesFromS3(wordFiles, bucketName)

	for i := range nodes {
		err = checkWorker(nodes[i])
		if err != nil {
			utility.CheckErrorNonFatal(err)
			nodes = append(nodes[:i], nodes[i+1:]...)
		}
	}

	for i := range nodes {
		err = callLoadTopologyOnWorker(nodes, nodes[i])
		if err != nil {
			return err
		}
	}

	err = callMapOnWorkers(words, nodes) //End of this function means Map is done on all nodes
	if err != nil {
		return err
	}

	err = callBarrierOnWorkers(nodes) //End of this function means results are ready
	if err != nil {
			return err
	}

	counted, err := getResultsOnWorkers(nodes)
	if err != nil {
		return err
	}
	saveResults(counted, "WordCount_Res")

	end := time.Now()
	diff := end.Sub(start)
	go logData(counted, diff, len(nodeConf.Workers))

	*res = true
	return nil
}

func logData(words []wordCountUtils.WordCount, latency time.Duration, workers int) {
	nWords := wordCountUtils.CountTotalWords(words)
	sec := latency.Seconds()
	throughput := float64(nWords) / sec
	myMetrics := metrics.WordCountMetrics{nWords, sec, throughput, workers}
	err:= metrics.AppendApplicationMetrics(metricsJsonPath, myMetrics)
	utility.CheckErrorNonFatal(err)
}

func main() {
	utility.ImportJson(nodesJsonPath, &nodeConf)

	flag.StringVar(&bucketName, "bucketName", "cesto93", "The rpc port of the master for the client")
	flag.Parse()
	fmt.Println("Starting rpc service on Master node on port " + nodeConf.MasterPort)
	master := new(Master)
	rpcUtils.ServRpc(nodeConf.MasterPort, "Master", master)
}
