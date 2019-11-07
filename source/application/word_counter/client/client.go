package main

import (
	"fmt"
	"io/ioutil"
	"flag"
	"log"
	"net/rpc"
	"progettoSDCC/source/application/word_counter/utility"
	"progettoSDCC/source/application/word_counter/storage"
	"progettoSDCC/source/application/word_counter/rpc_worker"
)

const(
	bucketName = "cesto93"
)

func putWordsToServer(names []string, paths []string){
	s := storage.New(bucketName)
	for i := range paths {
		file, err := ioutil.ReadFile(paths[i])
		if err != nil {
			fmt.Println("File reading error", err)
		}
		err = s.Write(names[i], []byte(file))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func requestWordCount(wordFiles []string, node rpc_worker.Node) []rpc_worker.WordCount{
	var res []rpc_worker.WordCount
	client, err := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()
	err = client.Call("Master.DoWordCount", wordFiles, &res)
	if err != nil {
		log.Fatal("Error in rpc_Map: ", err)
	}
	return res
}

func removeWordsFromServer(paths []string){
	//TODO
}

func main(){
	var names, paths utility.ArrayFlags
	var load bool
	var serverAddr rpc_worker.Node
	flag.BoolVar(&load, "load", false, "Specify the load file operation")
	flag.Var(&names, "names", "Name of file to upload.")
	flag.Var(&paths, "paths", "Path of file to upload.")
	flag.Var(&serverAddr, "serverAddr", "The server port for the rpc")
	flag.Parse()
	if (load) {
		if( len(names) != len(paths)){
			fmt.Println("Error, paths and names must have same dimension")
		}
		putWordsToServer(names, paths)
	} else {
		results := requestWordCount(names, serverAddr)
		for _, res := range results {
			fmt.Println(res.Word, res.Occurrence)
		}
	}
}
