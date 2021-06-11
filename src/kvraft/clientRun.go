package kvraft

import (
	"fmt"
	"net/rpc"
	"time"
)

func MakeClerkRun(servers []string) *Clerk {
	ck := new(Clerk)
	ck.serversRun = servers
	ck.serverNumber = len(servers)
	ck.clientId = time.Now().UnixNano()
	ck.Test = false
	ck.startUserInterface()
	return ck
}

//setup rpc
func call(rpcname string, server string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", ":"+server)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err != nil {
		//fmt.Println(err)
		return false
	}
	return true
}

func (ck *Clerk) startUserInterface() {
	for {
		fmt.Println("-----------------------------------------")
		fmt.Println(" 1. Append")
		fmt.Println(" 2. Get ")
		fmt.Println(" 3. Put")
		fmt.Print("-----------------------------------------\n\n")
		fmt.Print(" Select a option ->: ")
		var op string
		fmt.Scan(&op)
		if op == "3" {
			fmt.Print(" Enter Key: ")
			var key string
			fmt.Scan(&key)
			fmt.Print(" Enter Value: ")
			var value string
			fmt.Scan(&value)
			ck.Put(key, value)
		} else if op == "2" {
			fmt.Print(" Enter Key: ")
			var key string
			fmt.Scan(&key)
			v := ck.Get(key)
			fmt.Println("Key: ", key, ", Value: ", v)
		} else if op == "1" {
			fmt.Print(" Enter Key: ")
			var key string
			fmt.Scan(&key)
			fmt.Print(" Enter Value: ")
			var value string
			fmt.Scan(&value)
			ck.Append(key, value)
		} else {
			fmt.Println("INCORRECT  Reselect")
			continue
		}
	}
}
