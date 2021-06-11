package raft

// This file is for running the server

import (
	"errors"
	"fmt"
	"net/rpc"
	"strconv"
)

func MakeRun(rf *Raft, peersRun []string, me int, applyCh chan ApplyMsg) *Raft {
	rf.PeersRun = peersRun
	rf.Test = false
	rf.Network = Connect
	rf.PeerNumber = len(peersRun)
	setRaft(rf, me, applyCh)
	return rf
}

//wrapper for call to check network status
func (rf *Raft) call(rpcname string, server string, args interface{}, reply interface{}) bool {
	rf.mu.Lock()
	if rf.Network == Connect {
		rf.mu.Unlock()
		return call(rpcname, server, args, reply)
	} else {
		rf.mu.Unlock()
		return false
	}
}

//setup rpc
func call(rpcname string, server string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", server)
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

func (rf *Raft) HandleAppendEntriesRun(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	err := rf.checkNetwork()
	if err != nil {
		return err
	}
	rf.HandleAppendEntries(args, reply)
	return nil
}

func (rf *Raft) HandleRequestVoteRun(args *RequestVoteArgs, reply *RequestVoteReply) error {
	err := rf.checkNetwork()
	if err != nil {
		return err
	}
	rf.HandleRequestVote(args, reply)
	return nil
}

func (rf *Raft) checkNetwork() error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.Network == Disconnect {
		return errors.New("SERVER " + strconv.Itoa(rf.Me) + " DISCONNECT")
	}
	return nil
}

func (rf *Raft) Connect() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.Network = Connect
}

func (rf *Raft) Disconnect() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.Network = Disconnect
}

func (rf *Raft) PrintLog() {
	rf.mu.Lock()
	var state string
	if rf.State == Leader {
		state = "Leader"
	} else if rf.State == Cand {
		state = "Candidate"
	} else {
		state = "Follwer"
	}
	fmt.Println("-----------------------------------------")
	fmt.Println(state, " with ", rf.Term, " have LOG:")
	for _, vs := range rf.Log {
		fmt.Print("Command: ")
		fmt.Print(vs.Command)
		fmt.Println()
	}
	fmt.Print("-----------------------------------------\n\n")
	rf.mu.Unlock()
}
