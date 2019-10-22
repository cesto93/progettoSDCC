package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"word_counter/rpc_worker"
)

func import_json(path string, pointer interface{}) {
	file_json, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	err = json.Unmarshal(file_json, pointer)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func read_wordfiles(paths []string) []string {
	var texts []string
	for i := range paths {
		file, err := ioutil.ReadFile(paths[i])
		if err != nil {
			fmt.Println("File reading error", err)
			return nil
		}
		texts = append(texts, string(file))
	}
	return texts
}

//ASYNC
func call_map_on_workers(texts []string, nodes []rpc_worker.Node) {
	rpc_chan := make(chan *rpc.Call, len(nodes))
	for i := range texts {
		client, err := rpc.DialHTTP("tcp", "localhost:"+nodes[i % len(nodes)].Port)
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
func call_barrier_on_workers(nodes []rpc_worker.Node) {

	rpc_chan := make(chan *rpc.Call, len(nodes))
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", "localhost:"+nodes[i].Port)
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
func call_load_topology_on_worker(list_of_nodes []rpc_worker.Node, port string) {
	client, err := rpc.DialHTTP("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()
	err = client.Call("Worker.Load_Topology", list_of_nodes, nil)
	if err != nil {
		log.Fatal("Error in rpc_Map: ", err)
	}
}

//SYNC
func get_results_on_workers(nodes []rpc_worker.Node) []rpc_worker.Word_count {
	var res []rpc_worker.Word_count
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", "localhost:"+nodes[i].Port)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		defer client.Close()
		var counted []rpc_worker.Word_count
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

func main() {
	var word_files_list, workers_list string
	if len(os.Args) != 3 { //Parameter not inserted using default parameter
		//fmt.Println("Parameter not inserted using default parameter")
		word_files_list = "../word_files.json" 
		workers_list = "../workers.json"
	} else {
		word_files_list = os.Args[1] //the list of file to read
		workers_list = os.Args[2]    //the list of worker with it's rpc coordinate
	}	

	var nodes []rpc_worker.Node
	var word_files []string
	import_json(workers_list, &nodes)
	import_json(word_files_list, &word_files)

	s := read_wordfiles(word_files)

	for i := range nodes {
		call_load_topology_on_worker(nodes, nodes[i].Port)
	}
	call_map_on_workers(s, nodes) //End of this function means Map is done on all nodes

	call_barrier_on_workers(nodes) //End of this function means results are ready

	results := get_results_on_workers(nodes)

	for _, res := range results {
		fmt.Println(res.Word, res.Occurrence)
	}
}
