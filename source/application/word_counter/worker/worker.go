package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"word_counter/rpc_worker"
)

func serv_rpc(port string) {
	worker := new(rpc_worker.Worker)
	server := rpc.NewServer()
	err := server.RegisterName("Worker", worker)
	if err != nil {
		log.Fatal("Format of service rpc is not correct: ", err)
	}
	// Register an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	server.HandleHTTP("/", "/debug")

	// Listen for incoming messages on port
	l, e := net.Listen("tcp", ":"+port)
	if e != nil {
		log.Fatal("Listen error: ", e)
	}

	// Start go's http server on socket specified by l
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal("Serve error: ", err)
	}
}

func main() {

	port := os.Args[1]
	serv_rpc(port)
}
