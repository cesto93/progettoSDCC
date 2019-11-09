package main

import (
	"log"
	"net/rpc"
	"os"
	"sync"
	"fmt"
	"progettoSDCC/source/application/word_counter/word_count_utils"
	"progettoSDCC/source/application/word_counter/rpc_utils"
)

type Worker int

var nodes []rpc_utils.Node
var mapper_words []word_count_utils.WordCount
var reducer_words []word_count_utils.WordCount
var worker_state int
var mux sync.Mutex

const State_Idle = 0
const State_Map = 1
const State_Reducer = 2

func reducerKey(word string, n_nodes int) int {
	res := 0
	values := []byte(word)
	for i := 0; i < len(values); i++ {
		res += int(values[i])
	}
	return res % n_nodes
}

func shaffleAndSort(words []word_count_utils.WordCount, n_nodes int) [][]word_count_utils.WordCount {
	words_by_reducer := make([][]word_count_utils.WordCount, n_nodes)
	for i := range words_by_reducer {
		words_by_reducer[i] = make([]word_count_utils.WordCount, 0)
	}
	for i := range words {
		key := reducerKey(words[i].Word, n_nodes)
		words_by_reducer[key] = append(words_by_reducer[key], words[i])
	}
	return words_by_reducer
}

//ASYNC
func callReduce(words []word_count_utils.WordCount, nodes []rpc_utils.Node) {
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
	temp := word_count_utils.StringSplit(text)

	mux.Lock()
	if (mapper_words != nil) {
		for i := range mapper_words {
			temp = append(temp, mapper_words[i])
		}
	}
	mapper_words = word_count_utils.CountWords(temp) //We do a preliminary reduce
	mux.Unlock()

	*res = true
	return nil
}

func (t *Worker) Reduce(words []word_count_utils.WordCount, res *bool) error {
	worker_state = State_Reducer

	mux.Lock()
	if (reducer_words != nil) {
		for i := range reducer_words {
			words = append(words, reducer_words[i])
		}
	}
	reducer_words = word_count_utils.CountWords(words)
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

func (t *Worker) LoadTopology(nodes_list []rpc_utils.Node, res *bool) error {
	nodes = nodes_list

	//reset_words
	worker_state = State_Idle
	reducer_words = nil
	mapper_words = nil

	*res = true
	return nil
}

func (t *Worker) GetResults(state bool, res *[]word_count_utils.WordCount) error {
	if worker_state == State_Reducer {
		*res = reducer_words
		worker_state = State_Idle
	}
	return nil
}

func main() {
	port := os.Args[1]
	fmt.Println("Starting rpc service")
	worker := new(Worker)
	rpc_utils.ServRpc(port, "Worker", worker)
}
