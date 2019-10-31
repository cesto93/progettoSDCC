package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"flag"
	"progettoSDCC/source/application/word_counter/rpc_worker"
	"progettoSDCC/source/application/word_counter/utility"
	"progettoSDCC/source/application/word_counter/storage"
)

const(
	bucketName = "cesto93"
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

func read_wordfiles_fromS3(keys []string) []string {
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
func call_map_on_workers(texts []string, nodes []rpc_worker.Node) {
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
func call_barrier_on_workers(nodes []rpc_worker.Node) {

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
func call_load_topology_on_worker(topology []rpc_worker.Node, node rpc_worker.Node) {
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
func get_results_on_workers(nodes []rpc_worker.Node) []rpc_worker.Word_count {
	var res []rpc_worker.Word_count
	for i := range nodes {
		client, err := rpc.DialHTTP("tcp", nodes[i].Address + ":" + nodes[i].Port)
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

	var nodes rpc_worker.NodeList
	var word_files utility.ArrayFlags
	var localStorage bool
	var s []string

	flag.Var(&word_files, "files", "The file to request.")
	flag.Var(&nodes, "ports", "The list of worker with it's rpc coordinate")
	flag.BoolVar(&localStorage, "local-storage", false, "If the storage  of the file is local")
	flag.Parse()

	if (localStorage) {
		s = read_wordfiles(word_files)
	} else {
		s = read_wordfiles_fromS3(word_files)
	}

	for i := range nodes {
		call_load_topology_on_worker(nodes, nodes[i])
	}

	call_map_on_workers(s, nodes) //End of this function means Map is done on all nodes

	call_barrier_on_workers(nodes) //End of this function means results are ready

	results := get_results_on_workers(nodes)

	for _, res := range results {
		fmt.Println(res.Word, res.Occurrence)
	}
}
