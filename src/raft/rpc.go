package raft

const (
	Leader = iota + 1
	Cand
	Follwer
)
const (
	DidNotWin = iota + 1
	Win
)
const (
	Connect = iota + 1
	Disconnect
)
const (
	Ok = iota + 1
	Fail
)

const (
	Append = iota + 1
	CommitAndHeartBeat
	HeartBeat
)

type ApplyMsg struct {
	Command      interface{}
	CommandValid bool
	CommandIndex int
}

type Entry struct {
	Index   int
	Command interface{}
	Term    int
	Id      int
}

type ClientMessageArgs struct {
	Message int
}

type ClientMessageReply struct {
	Message int
	Status  int
}

type AppendEntriesArgs struct {
	Job          int
	Term         int
	LeaderId     int
	Entries      []Entry
	PrevLogIndex int //index of log entry immediately precedingnew ones
	PrevLogTerm  int //term of PrevLogIndex entry
	LeaderCommit int //leader’s commitIndex
}

type AppendEntriesReply struct {
	LastIndex int
	Term      int
	Success   bool
}

type RequestVoteArgs struct {
	PeerId       int
	Term         int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
	State       int
}
