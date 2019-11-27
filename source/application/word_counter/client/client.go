package main

import (
	"fmt"
	"io/ioutil"
	"flag"
	"log"
	"net/rpc"
	"progettoSDCC/source/utility"
	"progettoSDCC/source/application/word_counter/storage"
	"progettoSDCC/source/application/word_counter/rpcUtils"
	"progettoSDCC/source/application/word_counter/wordCountUtils"
)

func putWordsToServer(bucketName string, names []string, paths []string){
	s := storage.New(bucketName)
	for i := range paths {
		file, err := ioutil.ReadFile(paths[i])
		if err != nil {
			fmt.Println("File reading error", err)
		}
		err = s.Write(names[i], []byte(file))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func requestWordCount(wordFiles []string, node rpcUtils.Node) []wordCountUtils.WordCount{
	var res []wordCountUtils.WordCount
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

func removeWordsFromServer(bucketName string, keys []string){
	s := storage.New(bucketName)
	err := s.Delete(keys)
	if err != nil {
			log.Fatal(err)
	}
}

func listFileOnServer(bucketName string) []string {
	s := storage.New(bucketName)
	res, err := s.List()
	if err != nil {
			log.Fatal(err)
	}
	return res
}

func main(){
	var bucketName string
	var names, paths utility.ArrayFlags
	var load, delete, list, count bool
	var serverAddr rpcUtils.Node
	flag.BoolVar(&load, "load", false, "Specify the load file operation")
	flag.BoolVar(&delete, "delete", false, "Specify the delete file operation")
	flag.BoolVar(&list, "list", false, "Specify the list file operation")
	flag.BoolVar(&count, "count", false, "Specify the count word operation")

	flag.StringVar(&bucketName, "bucket", "cesto93", "The name of the bucket on aws")
	flag.Var(&names, "names", "Name of file to upload.")
	flag.Var(&paths, "paths", "Path of file to upload.")
	flag.Var(&serverAddr, "serverAddr", "The server port for the rpc")
	flag.Parse()
	if (load) {
		if( len(names) != len(paths)){
			log.Fatal("Error, paths and names must have same dimension")
		}
		putWordsToServer(bucketName, names, paths)
	} else if (delete) {
		removeWordsFromServer(bucketName, names)
	} else if (list) {
		fmt.Println(listFileOnServer(bucketName))
	} else if (count){
		fmt.Println("Requesting word count to master")
		results := requestWordCount(names, serverAddr)
		for _, res := range results {
			fmt.Println(res.Word, res.Occurrence)
		}
	}
}
