package raft

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func MakeRun(peersRun []string, me int, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.PeersRun = peersRun
	rf.Test = false
	rf.Network = Connect
	setRaft(rf, me, applyCh)
	rf.readPersistRun()
	return rf
}

func (rf *Raft) connect() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.Network = Connect
}

func (rf *Raft) disconnect() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.Network = Disconnect
}

func (rf *Raft) printLog() {
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
		fmt.Print(", Id: ")
		fmt.Print(vs.Id)
		fmt.Println()
	}
	fmt.Print("-----------------------------------------\n\n")
	rf.mu.Unlock()
}

func (rf *Raft) readPersistRun() {
	fileName := strconv.Itoa(rf.Me) + ".yys"
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(fileName, " is not created")
		return
	}
	d := json.NewDecoder(file)
	var CurrentTerm int
	var VotedFor int
	var Logs []Entry
	d.Decode(&CurrentTerm)
	d.Decode(&VotedFor)
	d.Decode(&Logs)
	rf.Term = CurrentTerm
	rf.VotedFor = VotedFor
	rf.Log = Logs
	fmt.Println("\nREAD FROM DISK--------------")
	fmt.Println("Term: ", rf.Term)
	fmt.Println("VotedFor", rf.VotedFor)
	for _, vs := range rf.Log {
		fmt.Print("Command: ")
		fmt.Print(vs.Command)
		fmt.Print(", Id: ")
		fmt.Print(vs.Id)
		fmt.Println()
	}
	fmt.Println("----------------------------")
}
