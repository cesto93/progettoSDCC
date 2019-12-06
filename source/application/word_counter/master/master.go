package main

import (
	"fmt"
	"log"
	"net/rpc"
	"flag"
	"progettoSDCC/source/application/word_counter/storage"
	"progettoSDCC/source/application/word_counter/rpcUtils"
	"progettoSDCC/source/application/word_counter/wordCountUtils"
	"progettoSDCC/source/utility"
)

type Master int

var nodeConf rpcUtils.NodeConfiguration
var bucketName string

const (
	nodesJsonPath = "../configuration/generated/app_node.json"
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
func callMapOnWorkers(texts []string, nodes []rpcUtils.Node) {
	rpcChan := make(chan *rpc.Call, len(nodes))
	for i := range texts {
		client, err := rpc.DialHTTP("tcp", nodes[i % len(nodes)].Address + ":" + nodes[i % len(nodes)].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		client.Go("Worker.Map", texts[i], nil, rpcChan)
	}
	for i := range texts {
		divCall := <-rpcChan
		if divCall.Error != nil {
			log.Fatal("Error in rpcMap num ", i % len(nodes), " :", divCall.Error.Error())
		}
	}
}

//ASYNC
func callBarrierOnWorkers(nodes []rpcUtils.Node) {

	rpcChan := make(chan *rpc.Call, len(nodes))
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		state := true
		client.Go("Worker.EndMapFase", state, nil, rpcChan)
	}
	for i := range nodes {
		divCall := <-rpcChan
		if divCall.Error != nil {
			log.Fatal("Error in rpcEndMapFase num ", i, " :", divCall.Error.Error())
		}
	}
}

//SYNC
func callLoadTopologyOnWorker(topology []rpcUtils.Node, node rpcUtils.Node) {
	client, err := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()
	err = client.Call("Worker.LoadTopology", topology, nil)
	if err != nil {
		log.Fatal("Error in rpcMap: ", err)
	}
}

//SYNC
func getResultsOnWorkers(nodes []rpcUtils.Node) []wordCountUtils.WordCount {
	var res []wordCountUtils.WordCount
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		var counted []wordCountUtils.WordCount
		state := true
		err = client.Call("Worker.GetResults", state, &counted)
		if err != nil {
			log.Fatal("Error in rpcMap: ", err)
		}
		//fmt.Println("words by reducer ", i, " = ", len(counted))
		for j := range counted {
			res = append(res, counted[j])
		}
	}
	return res
}

func (t *Master) DoWordCount(wordFiles []string, res *bool) error{
	fmt.Println("Request received")
	nodes := nodeConf.Workers

	s := readWordfilesFromS3(wordFiles, bucketName)

	for i := range nodes {
		callLoadTopologyOnWorker(nodes, nodes[i])
	}

	callMapOnWorkers(s, nodes) //End of this function means Map is done on all nodes

	callBarrierOnWorkers(nodes) //End of this function means results are ready

	counted := getResultsOnWorkers(nodes)
	saveResults(counted, "WordCount_Res")
	*res = true
	return nil
}

func main() {

	//flag.Var(&nodes, "workerAddr", "The list of worker with it's rpc coordinate")
	//flag.StringVar(&masterPort, "masterPort", "1049", "The rpc port of the master for the client")
	utility.ImportJson(nodesJsonPath, &nodeConf)

	flag.StringVar(&bucketName, "bucketName", "cesto93", "The rpc port of the master for the client")
	flag.Parse()
	fmt.Println("Starting rpc service on Master node on port " + nodeConf.MasterPort)
	master := new(Master)
	rpcUtils.ServRpc(nodeConf.MasterPort, "Master", master)
}
