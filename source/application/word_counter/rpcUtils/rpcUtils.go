package rpcUtils

import (
	"net/rpc"
	"net/http"
	"log"
	"net"
)

type Node struct {
	Address string
	Port string
}

func ServRpc(port string, name string, rcvr interface{}) {
	server := rpc.NewServer()
	err := server.RegisterName(name, rcvr)
	if err != nil {
		log.Fatal("Format of service rpc is not correct: ", err)
	}
	// Register an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	server.HandleHTTP("/", "/debug")
	l, e := net.Listen("tcp", ":" + port)
	if e != nil {
		log.Fatal("Listen error: ", e)
	}
	err = http.Serve(l, nil) // Start go's http server on socket specified by l
	if err != nil {
		log.Fatal("Serve error: ", err)
	}
}