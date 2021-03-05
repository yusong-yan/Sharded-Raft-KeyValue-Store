package raft

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.Peers[server].Call("Raft.HandleRequestVote", args, reply)
	return ok
}
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.Peers[server].Call("Raft.HandleAppendEntries", args, reply)
	return ok
}

func (rf *Raft) HandleAppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term >= rf.Term {
		//heartBeat
		rf.Term = args.Term
		if rf.IsLeader {
			rf.BecomeFollwerFromLeader <- true
		} else {
			rf.ReceiveHB <- true
		}
		rf.setFollwer()
		reply.Success = true
		reply.Term = rf.Term

		if args.Job == Append {
			//APPEND
			entr, find := rf.getLogAtIndexWithoutLock(args.PrevLogIndex)
			if !find { //give the last index
				reply.LastIndex = rf.getLastLogEntryWithoutLock().Index
				reply.Success = false
			} else {
				if entr.Term != args.PrevLogTerm {
					rf.Log = rf.Log[:indexInLog(args.PrevLogIndex)]
					reply.LastIndex = -1
					reply.Success = false
				} else {
					rf.Log = rf.Log[:indexInLog(args.PrevLogIndex+1)]
					rf.Log = append(rf.Log, args.Entries...)
					if (len(rf.Log)) > 0 {
						rf.Log[len(rf.Log)-1].Term = args.Term
					}
					reply.LastIndex = -1
					reply.Success = true
				}
			}
		} else if args.Job == CommitAndHeartBeat {
			rf.CommitIndex = min(args.LeaderCommit, rf.getLastLogEntryWithoutLock().Index)
			rf.startApply(rf.CommitIndex)
		}
		rf.persist()
		return
	}
	//TERM IS BIGGER JUST REPLY TERM
	reply.Term = rf.Term
	reply.Success = false
	rf.persist()
	return
}

func (rf *Raft) HandleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rfLastIndex := rf.getLastLogEntryWithoutLock().Index
	rfLastTerm := rf.termForLog(rfLastIndex)

	if args.Term > rf.Term {
		rf.Term = args.Term
		if rf.IsLeader {
			rf.BecomeFollwerFromLeader <- true
		} else {
			rf.ReceiveHB <- true
		}
		rf.setFollwer()
		rf.VotedFor = -1
	}
	if (rf.VotedFor == -1) && ((rfLastTerm < args.LastLogTerm) || ((rfLastTerm == args.LastLogTerm) && (rfLastIndex <= args.LastLogIndex))) {
		rf.VotedFor = args.PeerId
		reply.VoteGranted = true
	} else {
		reply.VoteGranted = false
	}
	reply.Term = rf.Term
	reply.State = rf.State
	rf.persist()
}
