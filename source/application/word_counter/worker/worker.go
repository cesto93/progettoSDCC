package main

import (
	"net/rpc"
	"os"
	"sync"
	"fmt"
	"time"
	"progettoSDCC/source/application/word_counter/wordCountUtils"
	"progettoSDCC/source/application/word_counter/rpcUtils"
	"progettoSDCC/source/utility"
	"progettoSDCC/source/appMetrics"
)

type Worker int

var nodes []rpcUtils.Node
var mapper_words []wordCountUtils.WordCount
var reducer_words []wordCountUtils.WordCount
var worker_state int
var mux sync.Mutex

var index int
var port string

const (
	State_Idle = 0
	State_Map = 1
	State_Reducer = 2
	portJsonPath = "../configuration/generated/port.json"
	metricsJsonPath = "../log/app_metrics.json"
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

func (t *Worker) Map(text string, res *bool) error {
	
	//LOG APP METRICS
	start := time.Now()
	//END LOG APP METRICS
	
	fmt.Println("Received map request\n")
	if worker_state == State_Reducer {
		fmt.Println("busy\n")
	}
	worker_state = State_Map
	temp := wordCountUtils.StringSplit(text)

	mux.Lock()
	if (mapper_words != nil) {
		for i := range mapper_words {
			temp = append(temp, mapper_words[i])
		}
	}
	temp = wordCountUtils.CountWords(temp) //We do a preliminary reduce
	mapper_words = temp
	mux.Unlock()
	
	//LOG APP METRICS
	end := time.Now()
	diff := end.Sub(start)
	go logWorkerData(temp, diff, fmt.Sprintf("%d", index))
	//END LOG APP METRICS

	*res = true
	return nil
}

func (t *Worker) LoadReducerWords(words []wordCountUtils.WordCount, res *bool) error {
	worker_state = State_Reducer

	mux.Lock()
	if (reducer_words != nil) {
		words = append(words, reducer_words...)
	}
	reducer_words = words
	mux.Unlock()

	*res = true
	return nil
}

//ASYNC
func (t *Worker) EndMapFase(state bool, res *bool) error {
	words_by_reducer := shaffleAndSort(mapper_words, len(nodes))
	rpc_chan := make(chan *rpc.Call, len(nodes))

	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
		if err != nil {
			return fmt.Errorf("Error in dialing: %v", err)
		}
		defer client.Close()
		client.Go("Worker.LoadReducerWords", words_by_reducer[i], nil, rpc_chan)
	}
	for i := range nodes {
		divCall := <-rpc_chan
		if divCall.Error != nil {
			return fmt.Errorf("Error in rpc_Reduce num %d: %v", i, " :", divCall.Error.Error())
		}
	}

	*res = true
	return nil
}

func (t *Worker) LoadTopology(nodes_list []rpcUtils.Node, res *bool) error {
	mux.Lock()
	nodes = nodes_list
	worker_state = State_Idle
	reducer_words = nil
	mapper_words = nil
	mux.Unlock()
	
	for i:= range nodes {
		if nodes[i].Port == port {
			index = i
			break
		}
	}
	*res = true
	return nil
}

func (t *Worker) CheckConn(arg bool, res *bool) error {
	*res = true
	return nil
}

func (t *Worker) GetResults(state bool, res *[]wordCountUtils.WordCount) error {
	fmt.Println("Received reduce request\n")
	if worker_state == State_Reducer {
		*res = wordCountUtils.CountWords(reducer_words)
		mux.Lock()
		worker_state = State_Idle
		reducer_words = nil
		mapper_words = nil
		mux.Unlock()
	}
	return nil
}

//LOG APP METRICS
func logWorkerData(words []wordCountUtils.WordCount, latency time.Duration, workerId string) {
	nWords := wordCountUtils.CountTotalWords(words)
	sec := latency.Seconds()
	throughput := float64(nWords) / sec
	labels := []string{"WordElaborated", "ElaborationTime", "ThroughPut"}
	values := []interface{}{nWords, sec, throughput}
	myMetrics:= appMetrics.NewAppMetrics("WordCount", "Mapper_" + workerId, labels, values)
	err:= appMetrics.AppendApplicationMetrics(metricsJsonPath, myMetrics)
	utility.CheckErrorNonFatal(err)
}
//END LOG APP METRICS

func main() {
	var err error
	if len(os.Args) == 2 {
		port = os.Args[1]
		utility.CheckError(err)
	} else {
		utility.ImportJson(portJsonPath, &port)
 		utility.CheckError(err)
	}
	
	utility.CheckError(err)
	fmt.Println("Starting rpc service on worker node at port " + port)
	worker := new(Worker)
	rpcUtils.ServRpc(port, "Worker", worker)
}
