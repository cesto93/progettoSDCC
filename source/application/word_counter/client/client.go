package main

import (
	"fmt"
	"io/ioutil"
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
		s.Write(names[i], []byte(file))
	}
}

func removeWordsFromServer(paths []string){
	//TODO
}

/*func requestWordCount(paths []string) []string{

}*/