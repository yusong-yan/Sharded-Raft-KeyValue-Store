package kvraft

// This file is for running the server

import (
	"Sharded-RaftKV/src/labgob"
	"Sharded-RaftKV/src/raft"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

func StartKVServerRun(servers []string, me int, maxraftstate int) *KVServer {
	labgob.Register(Op{})
	kv := new(KVServer)
	rf := new(raft.Raft)
	kv.Rf = rf
	kv.setup(servers[me])
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.applyCh = make(chan raft.ApplyMsg, 1)
	kv.Rf = raft.MakeRun(rf, servers, me, kv.applyCh)
	kv.readPersistRun(servers[me])
	kv.storage = map[string]string{}
	kv.idTable = map[int64]map[int64]Op{}
	go kv.listenApplyCh()
	kv.startUserInterface()
	return kv
}

func (kv *KVServer) setup(port string) {
	//setup my own server
	conn, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Listen:", err)
	}
	err = rpc.RegisterName("KVServer", kv)
	if err != nil {
		log.Fatal("KVServer:", err)
	}
	err = rpc.RegisterName("Raft", kv.Rf)
	if err != nil {
		log.Fatal("Raft:", err)
	}
	rpc.HandleHTTP()
	fmt.Println("\nSETUP KVSERVER DONE")
	go http.Serve(conn, nil)
}

func (kv *KVServer) PrintStateMachine() {
	kv.mu.Lock()
	fmt.Println("-----------------------------------------")
	fmt.Println("State Machine:")
	for i, vs := range kv.storage {
		fmt.Print("Key: ", i)
		fmt.Println(", Value: ", vs)
	}
	fmt.Print("-----------------------------------------\n\n")
	kv.mu.Unlock()
}

func (kv *KVServer) startUserInterface() {
	for {
		//fmt.Println("-----------------------------------------")
		fmt.Println("\n a.PrintLog")
		fmt.Println(" b.Connect ")
		fmt.Println(" c.Disconnect")
		fmt.Println(" d.PrintStateMachine")
		//fmt.Println("-----------------------------------------\n")
		fmt.Print(" Enter: ")
		var op string
		fmt.Scan(&op)
		if op == "b" {
			kv.Rf.Connect()
		} else if op == "c" {
			kv.Rf.Disconnect()
		} else if op == "a" {
			kv.Rf.PrintLog()
		} else if op == "d" {
			kv.PrintStateMachine()
		} else {
			fmt.Println("INCORRECT  Reselect")
			continue
		}
	}
}

func (kv *KVServer) GetRun(args *GetArgs, reply *GetReply) error {
	kv.Get(args, reply)
	return nil
}

func (kv *KVServer) PutAppendRun(args *PutAppendArgs, reply *PutAppendReply) error {
	kv.PutAppend(args, reply)
	return nil
}

func (kv *KVServer) readPersistRun(server string) {
	fileName := server + ".yys"
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(fileName, " is not created")
		return
	}
	d := json.NewDecoder(file)
	var CurrentTerm int
	var VotedFor int
	var len int
	Logs := []raft.Entry{}
	d.Decode(&CurrentTerm)
	d.Decode(&VotedFor)
	d.Decode(&len)
	for i := 0; i < len; i++ {
		var op Op
		var id int
		var index int
		var term int
		d.Decode(&op)
		d.Decode(&id)
		d.Decode(&index)
		d.Decode(&term)
		e := raft.Entry{}
		e.Command = op
		e.Id = id
		e.Index = index
		e.Term = term
		Logs = append(Logs, e)
	}
	kv.Rf.Term = CurrentTerm
	kv.Rf.VotedFor = VotedFor
	kv.Rf.Log = Logs
	fmt.Println("\nREAD FROM DISK--------------")
	fmt.Println("Term: ", kv.Rf.Term)
	fmt.Println("VotedFor", kv.Rf.VotedFor)
	for _, vs := range kv.Rf.Log {
		fmt.Print("Command: ")
		fmt.Print(vs.Command)
		fmt.Println()
	}
	fmt.Println("----------------------------")
}
