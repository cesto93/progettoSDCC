package rpc_worker

import (
	"log"
	"net/rpc"
	"strings"
	"unicode"
	"sync"
)

type Node struct {
	Address string
	Port string
}

type Word_count struct {
	Word       string
	Occurrence int
}
type Worker int

var nodes []Node
var mapper_words []Word_count
var reducer_words []Word_count
var worker_state int
var mux sync.Mutex

const State_Idle = 0
const State_Map = 1
const State_Reducer = 2

func string_split(text string) []Word_count {
	text = strings.ToLower(text)
	text = strings.Replace(text, "\n", " ", -1)
	words := strings.Split(text, " ")
	
	var counted []Word_count
	for i := range words {
		 trimmed_word := strings.TrimFunc(words[i], func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		if trimmed_word != "" { 
			counted = append(counted, Word_count{trimmed_word, 1}) 
		}
	}
	return counted
}

func count_words(words []Word_count) []Word_count {
	var counted []Word_count
	for i := range words {
		var j int
		for j = 0; j < len(counted); j++ {
			if words[i].Word == counted[j].Word {
				counted[j].Occurrence += words[j].Occurrence
				break
			}
		}
		if j == len(counted) {
			counted = append(counted, words[i])
		}
	}
	return counted
}

func reducer_key(word string, n_nodes int) int {
	res := 0
	values := []byte(word)
	for i := 0; i < len(values); i++ {
		res += int(values[i])
	}
	return res % n_nodes
}

func shaffle_and_sort(words []Word_count, n_nodes int) [][]Word_count {
	words_by_reducer := make([][]Word_count, n_nodes)
	for i := range words_by_reducer {
		words_by_reducer[i] = make([]Word_count, 0)
	}
	for i := range words {
		key := reducer_key(words[i].Word, n_nodes)
		words_by_reducer[key] = append(words_by_reducer[key], words[i])
	}
	return words_by_reducer
}

//ASYNC
func call_reduce(words []Word_count, nodes []Node) {
	words_by_reducer := shaffle_and_sort(words, len(nodes))
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
	temp := string_split(text)

	mux.Lock()
	if (mapper_words != nil) {
		for i := range mapper_words {
			temp = append(temp, mapper_words[i])
		}
	}
	mapper_words = count_words(temp) //We do a preliminary reduce
	mux.Unlock()

	*res = true
	return nil
}

func (t *Worker) Reduce(words []Word_count, res *bool) error {
	worker_state = State_Reducer

	mux.Lock()
	if (reducer_words != nil) {
		for i := range reducer_words {
			words = append(words, reducer_words[i])
		}
	}
	reducer_words = count_words(words)
	mux.Unlock()

	*res = true
	return nil
}

func (t *Worker) End_Map_Fase(state bool, res *bool) error {
	state = true
	call_reduce(mapper_words, nodes)
	*res = true
	return nil
}

func (t *Worker) Load_Topology(nodes_list []Node, res *bool) error {
	nodes = nodes_list

	//reset_words
	worker_state = State_Idle
	reducer_words = nil
	mapper_words = nil

	*res = true
	return nil
}

func (t *Worker) Get_Results(state bool, res *[]Word_count) error {
	if worker_state == State_Reducer {
		*res = reducer_words
		worker_state = State_Idle
	}
	return nil
}
