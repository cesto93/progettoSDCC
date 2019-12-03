package main

import (
	"log"
	"net/rpc"
	"os"
	"sync"
	"fmt"
	"strconv"
	"progettoSDCC/source/application/word_counter/wordCountUtils"
	"progettoSDCC/source/application/word_counter/rpcUtils"
	"progettoSDCC/source/utility"
)

type Worker int

var nodes []rpcUtils.Node
var mapper_words []wordCountUtils.WordCount
var reducer_words []wordCountUtils.WordCount
var worker_state int
var mux sync.Mutex

const (
	State_Idle = 0
	State_Map = 1
	State_Reducer = 2
	nodesJsonPath = "../configuration/generated/app_node.json"
	idWorkerPath = "../configuration/generated/id_worker.json"
)

func reducerKey(word string, n_nodes int) int {
	res := 0
	values := []byte(word)
	for i := 0; i < len(values); i++ {
		res += int(values[i])
	}
	return res % n_nodes
}

func shaffleAndSort(words []wordCountUtils.WordCount, n_nodes int) [][]wordCountUtils.WordCount {
	words_by_reducer := make([][]wordCountUtils.WordCount, n_nodes)
	for i := range words_by_reducer {
		words_by_reducer[i] = make([]wordCountUtils.WordCount, 0)
	}
	for i := range words {
		key := reducerKey(words[i].Word, n_nodes)
		words_by_reducer[key] = append(words_by_reducer[key], words[i])
	}
	return words_by_reducer
}

//ASYNC
func callReduce(words []wordCountUtils.WordCount, nodes []rpcUtils.Node) {
	words_by_reducer := shaffleAndSort(words, len(nodes))
	rpc_chan := make(chan *rpc.Call, len(nodes))
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		client.Go("Worker.Reduce", words_by_reducer[i], nil, rpc_chan)
	}
	for i := range nodes {
		divCall := <-rpc_chan
		if divCall.Error != nil {
			log.Fatal("Error in rpc_Reduce num ", i, " :", divCall.Error.Error())
		}
	}
}

func (t *Worker) Map(text string, res *bool) error {
	worker_state = State_Map
	temp := wordCountUtils.StringSplit(text)

	mux.Lock()
	if (mapper_words != nil) {
		for i := range mapper_words {
			temp = append(temp, mapper_words[i])
		}
	}
	mapper_words = wordCountUtils.CountWords(temp) //We do a preliminary reduce
	mux.Unlock()

	*res = true
	return nil
}

func (t *Worker) Reduce(words []wordCountUtils.WordCount, res *bool) error {
	worker_state = State_Reducer

	mux.Lock()
	if (reducer_words != nil) {
		for i := range reducer_words {
			words = append(words, reducer_words[i])
		}
	}
	reducer_words = wordCountUtils.CountWords(words)
	mux.Unlock()

	*res = true
	return nil
}

func (t *Worker) EndMapFase(state bool, res *bool) error {
	state = true
	callReduce(mapper_words, nodes)
	*res = true
	return nil
}

func (t *Worker) LoadTopology(nodes_list []rpcUtils.Node, res *bool) error {
	nodes = nodes_list

	//reset_words
	worker_state = State_Idle
	reducer_words = nil
	mapper_words = nil

	*res = true
	return nil
}

func (t *Worker) GetResults(state bool, res *[]wordCountUtils.WordCount) error {
	if worker_state == State_Reducer {
		*res = reducer_words
		worker_state = State_Idle
	}
	return nil
}

func main() {
	var nodeConf rpcUtils.NodeConfiguration
	var index int
	var err error
	if len(os.Args) == 2 {
		index, err = strconv.Atoi(os.Args[1])
		utility.CheckError(err)
	} else {
		err = utility.ImportJson(idWorkerPath, &index)
 		utility.CheckError(err)
	}
	utility.ImportJson(nodesJsonPath, &nodeConf)
	utility.CheckError(err)
	fmt.Println("Starting rpc service on worker node at port " + nodeConf.Workers[index].Port)
	worker := new(Worker)
	rpcUtils.ServRpc(nodeConf.Workers[index].Port, "Worker", worker)
}
