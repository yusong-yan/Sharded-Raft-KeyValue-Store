package raft

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"Sharded-RaftKV/src/labgob"
	"Sharded-RaftKV/src/labrpc"
)

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	Peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	Me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()
	//
	Log                     []Entry
	IsLeader                bool
	State                   int
	Term                    int
	VotedFor                int
	ReceiveHB               chan bool
	BecomeFollwerFromLeader chan bool
	NextIndex               map[int]int
	MatchIndex              map[int]int
	PeerAlive               map[int]bool
	OpenCommit              map[int]bool
	CommitIndex             int
	LastApply               int
	ApplyChan               chan ApplyMsg
	ApplyBuffer             chan bool
	PeerNumber              int
	Test                    bool     // for run
	PeersRun                []string // for run
	Network                 int      //for run
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	//test will call this make
	rf := &Raft{}
	rf.Peers = peers
	rf.PeerNumber = len(peers)
	rf.Test = true
	rf.persister = persister
	setRaft(rf, me, applyCh)

	rf.readPersist(persister.ReadRaftState())
	return rf
}

func setRaft(rf *Raft, me int, applyCh chan ApplyMsg) {
	rf.Me = me
	rf.State = Follwer
	rf.Log = []Entry{}
	rf.VotedFor = -1
	rf.IsLeader = false
	rf.Me = me
	rf.Term = 0
	rf.ReceiveHB = make(chan bool, 1)
	rf.BecomeFollwerFromLeader = make(chan bool, 1)
	rf.ApplyBuffer = make(chan bool, 1)
	rf.NextIndex = map[int]int{}
	rf.MatchIndex = map[int]int{}
	rf.PeerAlive = map[int]bool{}
	rf.OpenCommit = map[int]bool{}
	rf.ApplyChan = applyCh
	rf.CommitIndex = 0
	rf.LastApply = 0
	go rf.startElection()
}

func generateTime() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	diff := 700 - 350
	return 350 + r.Intn(diff)
}

func (rf *Raft) startElection() {
	for !rf.killed() {
		ticker := time.NewTicker(time.Duration(generateTime()) * time.Millisecond)
		electionResult := make(chan int, 1)
	Loop:
		for !rf.killed() {
			select {
			case <-ticker.C:
				//no leader, join election during the intervalTime
				interval := generateTime()
				ticker = time.NewTicker(time.Duration(interval) * time.Millisecond)
				go func() {
					electionResult <- rf.startAsCand(interval)
				}()
			case <-rf.ReceiveHB:
				//reset ticker since there is a leader
				ticker = time.NewTicker(time.Duration(generateTime()) * time.Millisecond)
			case a := <-electionResult:
				//if Win, break the follwer loop and become a leader
				if a == Win {
					break Loop
				}
			default:
			}
		}

		ticker.Stop()
		go rf.startAsLeader()
		//wait to become a follwer
		<-rf.BecomeFollwerFromLeader
	}
}

func (rf *Raft) startAsCand(interval int) int {
	cond := sync.NewCond(&rf.mu)
	// cand only a allow to become leader during the interval time. If it take longer
	// than interval time, this cand shouldn't be leader
	var needReturn bool
	needReturn = false
	go func(needReturn *bool, cond *sync.Cond) {
		time.Sleep(time.Duration(interval-20) * time.Millisecond)
		rf.mu.Lock()
		*needReturn = true
		cond.Signal()
		rf.mu.Unlock()
	}(&needReturn, cond)

	//setup args and rf
	hearedBack := 1
	votes := 1
	args := RequestVoteArgs{}
	rf.mu.Lock()
	rf.State = Cand
	rf.Term = rf.Term + 1
	rf.VotedFor = rf.Me
	args.Term = rf.Term
	args.PeerId = rf.Me
	args.LastLogIndex = rf.getLastLogEntryWithoutLock().Index
	args.LastLogTerm = rf.termForLog(args.LastLogIndex)
	rf.persist()
	rf.mu.Unlock()
	for s := 0; s < rf.PeerNumber; s++ {
		server := s
		if server == rf.Me {
			continue
		}
		reply := RequestVoteReply{}

		go func() {
			ok := rf.sendRequestVote(server, &args, &reply)
			//if the rpc couldn't reach that server or pass interval time, just end this thread
			if !ok || needReturn {
				rf.mu.Lock()
				hearedBack++
				cond.Signal()
				rf.mu.Unlock()
				return
			}
			rf.mu.Lock()
			hearedBack++
			if reply.Term > rf.Term && rf.State != Follwer {
				// if other node has high term, and you are still not follwer, become a follwer
				rf.ReceiveHB <- true
				rf.setFollwer()
				rf.Term = reply.Term
				rf.persist()
				cond.Signal()
				rf.mu.Unlock()
				return
			}

			if reply.VoteGranted == true && rf.State == Cand {
				votes++
			}
			cond.Signal()
			rf.mu.Unlock()
		}()
	}
	//wait.
	rf.mu.Lock()
	for hearedBack != rf.PeerNumber && votes <= rf.PeerNumber/2 && needReturn == false && rf.State == Cand {
		cond.Wait()
	}
	//decide
	if votes > rf.PeerNumber/2 && rf.State == Cand && needReturn == false {
		rf.mu.Unlock()
		return Win
	} else {
		rf.mu.Unlock()
		return DidNotWin
	}
}

func (rf *Raft) startAsLeader() {
	rf.mu.Lock()
	for i := 0; i < rf.PeerNumber; i++ {
		server := i
		// assume all the peer servers are dead, so that after the first heartbeat, leader
		// can start StartOnePeerAppend to each servers to make other servers have the same
		// log as this leader. It will also fix MatchIndex.
		// OpenCommit set false, because only after StartOnePeerAppend, that server is open to commit
		rf.NextIndex[server] = rf.getLastLogEntryWithoutLock().Index + 1
		rf.MatchIndex[server] = 0
		rf.PeerAlive[server] = false
		rf.OpenCommit[server] = false
		if (len(rf.Log)) > 0 {
			// for figure 8
			rf.Log[len(rf.Log)-1].Term = rf.Term
		}
	}
	rf.setLeader()
	rf.mu.Unlock()
	//rf.Start(nil)
	for !rf.killed() {
		go rf.sendHeartBeat()
		if rf.getState() != Leader {
			return
		}
		time.Sleep(time.Duration(160) * time.Millisecond)
	}
}

func (rf *Raft) sendHeartBeat() {
	if rf.getState() == Leader {
		for s := 0; s < rf.PeerNumber; s++ {
			server := s
			if server == rf.Me {
				continue
			}

			args := AppendEntriesArgs{}
			args.LeaderId = rf.Me
			args.Entries = []Entry{}
			args.Job = CommitAndHeartBeat
			rf.mu.Lock()
			args.LeaderCommit = rf.CommitIndex
			args.Term = rf.Term
			args.Job = HeartBeat
			if rf.OpenCommit[server] {
				// commit all the log for this server
				args.Job = CommitAndHeartBeat
			}
			rf.mu.Unlock()

			reply := AppendEntriesReply{}

			go func() {
				ok := rf.sendAppendEntries(server, &args, &reply)
				//Handle Reply
				if !ok {
					//if leader couldn't take to this machine, mark it as dead
					//so it will save sometime for StartPeerAppend, because
					//StartPeerAppend will not sendAppendEntries to this machine
					//until sendHeartbeat detect this server is alive again
					rf.mu.Lock()
					rf.PeerAlive[server] = false
					rf.OpenCommit[server] = false
					rf.mu.Unlock()
					return
				}
				rf.mu.Lock()
				//become follwer is term is smaller
				if reply.Term > rf.Term && rf.State == Leader {
					rf.Term = reply.Term
					rf.BecomeFollwerFromLeader <- true
					rf.setFollwer()
					rf.persist()
					rf.mu.Unlock()
					return
				}
				//Now, you realize this server is back online, so do somthing to it
				//Mark it alive and startOnePeerAppend to fix it log, then make it
				//will also make it open to commit
				if !rf.PeerAlive[server] && rf.State == Leader {
					rf.PeerAlive[server] = true
					go func() {
						rf.StartOnePeerAppend(server)
					}()
				}
				rf.mu.Unlock()
			}()
		}
	}
}

func (rf *Raft) Start(Command interface{}) (int, int, bool) {
	Index := -1
	Term := -1
	IsLeader := rf.getState() == Leader
	//check if ID exist
	if IsLeader {
		hearedBack := 1
		cond := sync.NewCond(&rf.mu)
		rf.mu.Lock()
		Term = rf.Term
		newE := Entry{}
		newE.Command = Command
		newE.Index = rf.getLastLogEntryWithoutLock().Index + 1
		newE.Term = rf.Term
		rf.Log = append(rf.Log, newE)
		Index = rf.getLastLogEntryWithoutLock().Index
		rf.persist()
		rf.mu.Unlock()
		for i := 0; i < rf.PeerNumber; i++ {
			server := i
			if server == rf.Me {
				continue
			}
			go func() {
				rf.StartOnePeerAppend(server)
				rf.mu.Lock()
				hearedBack++
				cond.Signal()
				rf.mu.Unlock()
			}()
		}

		//wait
		rf.mu.Lock()
		for hearedBack != rf.PeerNumber && rf.CommitIndex < Index && rf.IsLeader {
			cond.Wait()
		}
		//if rf.CommitIndex>=Index, it means that more than half of the machine are approve
		//this commit

		//decide
		if !rf.IsLeader {
			Index, Term, IsLeader = -1, -1, false
		}
		rf.mu.Unlock()
	}
	return Index, Term, IsLeader
}

//fix one peer
func (rf *Raft) StartOnePeerAppend(server int) bool {
	result := false
	if rf.getState() == Leader {
		//set up sending log
		entries := []Entry{}
		rf.mu.Lock()
		for i := rf.MatchIndex[server] + 1; i <= rf.getLastLogEntryWithoutLock().Index; i++ {
			entry, find := rf.getLogAtIndexWithoutLock(i)
			if !find {
				entries = []Entry{}
				break
			}
			entries = append(entries, entry)
		}
		args := AppendEntriesArgs{}
		args.LeaderId = rf.Me
		args.Term = rf.Term
		args.PrevLogIndex = rf.MatchIndex[server]
		args.PrevLogTerm = rf.termForLog(args.PrevLogIndex)
		args.Entries = entries
		args.LeaderCommit = rf.CommitIndex
		args.Job = Append
		rf.mu.Unlock()
		for rf.getState() == Leader && !rf.killed() {
			reply := AppendEntriesReply{}
			rf.mu.Lock()
			if rf.PeerAlive[server] && rf.IsLeader {
				rf.mu.Unlock()
				ok := rf.sendAppendEntries(server, &args, &reply)
				if !ok {
					rf.mu.Lock()
					rf.PeerAlive[server] = false
					rf.OpenCommit[server] = false
					rf.mu.Unlock()
					result = false
					break
				}
			} else {
				// if it is not alive, just end this function
				rf.mu.Unlock()
				result = false
				break
			}

			if reply.Success {
				// It means that this server have the exact same log as leader server
				//update, make it open to commit so heartbeat will commit it
				rf.mu.Lock()
				rf.MatchIndex[server] = len(args.Entries) + args.PrevLogIndex
				rf.NextIndex[server] = rf.MatchIndex[server] + 1
				rf.OpenCommit[server] = true
				rf.PeerAlive[server] = true
				//check if there is over half machine approve the commit
				if rf.updateCommitForLeader() && rf.IsLeader {
					//if so, leader just commit it
					rf.startApply(rf.CommitIndex)
				}
				rf.mu.Unlock()
				result = true
				break
			} else {
				//resend
				rf.mu.Lock()
				rf.PeerAlive[server] = true
				args.Term = rf.Term
				args.LeaderCommit = rf.CommitIndex
				if reply.LastIndex != -1 {
					//if server's log size bigger than rflog size
					args.PrevLogIndex = reply.LastIndex
				} else {
					args.PrevLogIndex = args.PrevLogIndex - 1
				}
				args.PrevLogTerm = rf.termForLog(args.PrevLogIndex)
				args.Entries = rf.Log[indexInLog(args.PrevLogIndex+1):]
				rf.mu.Unlock()
			}
		}
	}
	return result
}

func (rf *Raft) updateCommitForLeader() bool {
	beginIndex := rf.CommitIndex + 1
	lastCommittedIndex := -1
	updated := false
	for ; beginIndex <= rf.getLastLogEntryWithoutLock().Index; beginIndex++ {
		granted := 1

		for Server, ServerMatchIndex := range rf.MatchIndex {
			if Server == rf.Me || !rf.PeerAlive[Server] {
				continue
			}
			if ServerMatchIndex >= beginIndex {
				granted++
			}
		}

		if granted >= rf.PeerNumber/2+1 {
			lastCommittedIndex = beginIndex
		}
	}
	if lastCommittedIndex > rf.CommitIndex && rf.IsLeader {
		rf.CommitIndex = lastCommittedIndex
		updated = true
	}
	return updated
}

func (rf *Raft) startApply(CommitIndex int) {
	for CommitIndex > rf.LastApply && !rf.killed() {
		rf.LastApply = rf.LastApply + 1
		am := ApplyMsg{}
		am.Command = rf.Log[indexInLog(rf.LastApply)].Command
		am.CommandIndex = rf.LastApply
		am.CommandValid = true
		rf.ApplyChan <- am
	}
}

func (rf *Raft) persist() {
	if rf.Test {
		w := new(bytes.Buffer)
		e := labgob.NewEncoder(w)
		e.Encode(rf.Term)
		e.Encode(rf.VotedFor)
		e.Encode(rf.Log)
		data := w.Bytes()
		rf.persister.SaveRaftState(data)
	} else {
		fileName := rf.PeersRun[rf.Me] + ".yys"
		ofile, _ := os.Create(fileName)
		e := json.NewEncoder(ofile)
		e.Encode(rf.Term)
		e.Encode(rf.VotedFor)
		e.Encode(len(rf.Log))
		for i := 0; i < len(rf.Log); i++ {
			e.Encode(rf.Log[i].Command)
			e.Encode(rf.Log[i].Id)
			e.Encode(rf.Log[i].Index)
			e.Encode(rf.Log[i].Term)
		}
	}
}

func (rf *Raft) readPersist(data []byte) {
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var CurrentTerm int
	var VotedFor int
	var Logs []Entry
	d.Decode(&CurrentTerm)
	d.Decode(&VotedFor)
	d.Decode(&Logs)
	rf.Term = CurrentTerm
	rf.VotedFor = VotedFor
	rf.Log = Logs
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}
