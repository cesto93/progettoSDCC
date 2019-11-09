package main

import (
	"fmt"
	"log"
	"net/rpc"
	"flag"
	"progettoSDCC/source/application/word_counter/storage"
	"progettoSDCC/source/application/word_counter/rpc_utils"
	"progettoSDCC/source/application/word_counter/word_count_utils"
)

type Master int

var nodes rpc_utils.NodeList
var bucketName string

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

//ASYNC
func callMapOnWorkers(texts []string, nodes []rpc_utils.Node) {
	rpc_chan := make(chan *rpc.Call, len(nodes))
	for i := range texts {
		client, err := rpc.DialHTTP("tcp", nodes[i % len(nodes)].Address + ":" + nodes[i % len(nodes)].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		client.Go("Worker.Map", texts[i], nil, rpc_chan)
	}
	for i := range texts {
		divCall := <-rpc_chan
		if divCall.Error != nil {
			log.Fatal("Error in rpc_Map num ", i % len(nodes), " :", divCall.Error.Error())
		}
	}
}

//ASYNC
func callBarrierOnWorkers(nodes []rpc_utils.Node) {

	rpc_chan := make(chan *rpc.Call, len(nodes))
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		state := true
		client.Go("Worker.End_Map_Fase", state, nil, rpc_chan)
	}
	for i := range nodes {
		divCall := <-rpc_chan
		if divCall.Error != nil {
			log.Fatal("Error in rpc_End_Map_Fase num ", i, " :", divCall.Error.Error())
		}
	}
}

//SYNC
func callLoadTopologyOnWorker(topology []rpc_utils.Node, node rpc_utils.Node) {
	client, err := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()
	err = client.Call("Worker.Load_Topology", topology, nil)
	if err != nil {
		log.Fatal("Error in rpc_Map: ", err)
	}
}

//SYNC
func getResultsOnWorkers(nodes []rpc_utils.Node) []word_count_utils.WordCount {
	var res []word_count_utils.WordCount
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		var counted []word_count_utils.WordCount
		state := true
		err = client.Call("Worker.Get_Results", state, &counted)
		if err != nil {
			log.Fatal("Error in rpc_Map: ", err)
		}
		//fmt.Println("words by reducer ", i, " = ", len(counted))
		for j := range counted {
			res = append(res, counted[j])
		}
	}
	return res
}

func (t *Master) DoWordCount(word_files []string, res *[]word_count_utils.WordCount) error{
	fmt.Println("Request received")

	s := readWordfilesFromS3(word_files, bucketName)

	for i := range nodes {
		callLoadTopologyOnWorker(nodes, nodes[i])
	}

	callMapOnWorkers(s, nodes) //End of this function means Map is done on all nodes

	callBarrierOnWorkers(nodes) //End of this function means results are ready

	*res = getResultsOnWorkers(nodes)
	return nil
}

func main() {
	var masterPort string

	flag.Var(&nodes, "workerAddr", "The list of worker with it's rpc coordinate")
	flag.StringVar(&masterPort, "masterPort", "1049", "The rpc port of the master for the client")
	flag.StringVar(&bucketName, "bucketName", "cesto93", "The rpc port of the master for the client")
	flag.Parse()
	fmt.Println("Starting rpc service")
	master := new(Master)
	rpc_utils.ServRpc(masterPort, "Master", master)
}
