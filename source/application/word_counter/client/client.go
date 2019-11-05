package main

import (
	"fmt"
	"io/ioutil"
	"flag"
	"progettoSDCC/source/application/word_counter/utility"
	"progettoSDCC/source/application/word_counter/storage"
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

/*func requestWordCount(paths []string) []string{
	//TODO implemnt call to servRPC in master.go
}*/

func removeWordsFromServer(paths []string){
	//TODO
}

func main(){
	var names, paths utility.ArrayFlags
	flag.Var(&names, "n", "Name of file to upload.")
	flag.Var(&paths, "p", "Path of file to upload.")
	flag.Parse()
	if(len(names) != len(paths)){
		fmt.Println("Error, paths and names must have same dimension")
	}
	putWordsToServer(names, paths)
}
