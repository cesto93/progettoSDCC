package rpcUtils

import (
	"errors"
	"strings"
)

type NodeList []Node

func (i *NodeList) String() string {
    return "my node representation"
}

func (i *NodeList) Set(value string) error {
	if len(*i) > 0 {
		return errors.New("interval flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		data := strings.Split(dt, ":")
		node := Node{data[0], data[1]}
		*i = append(*i, node)
	}
	return nil
}

func (i *Node) String() string {
    return "my node representation"
}

func (i *Node) Set(value string) error {
	data := strings.Split(value, ":")
	*i = Node{data[0], data[1]}
	return nil
}