package raft

func (rf *Raft) getState() int {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.State
}

func (rf *Raft) getTerm() int {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.Term
}

func (rf *Raft) setLeader() {
	rf.IsLeader = true
	rf.State = Leader
}

func (rf *Raft) setFollwer() {
	rf.State = Follwer
	rf.IsLeader = false
}

//for log
func (rf *Raft) getLogAtIndexWithoutLock(index int) (Entry, bool) {
	if index == 0 {
		return Entry{}, true
	} else if len(rf.Log) == 0 {
		return Entry{}, false
	} else if (index < -1) || (index > rf.getLastLogEntryWithoutLock().Index) {
		return Entry{}, false
	} else {
		localIndex := index - rf.Log[0].Index
		return rf.Log[localIndex], true
	}
}

func (rf *Raft) getLogAtIndex(index int) (Entry, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.getLogAtIndexWithoutLock(index)
}

func (rf *Raft) getLastLogEntryWithoutLock() Entry {
	entry := Entry{}
	if len(rf.Log) == 0 {
		entry.Term = 0
		entry.Index = 0
	} else {
		entry = rf.Log[len(rf.Log)-1]
	}
	return entry
}

func (rf *Raft) getLastLogEntry() Entry {
	entry := Entry{}
	rf.mu.Lock()
	entry = rf.getLastLogEntryWithoutLock()
	rf.mu.Unlock()
	return entry
}

func (rf *Raft) termForLog(index int) int {
	entry, ok := rf.getLogAtIndexWithoutLock(index)
	if ok {
		return entry.Term
	} else {
		return -1
	}
}

func indexInLog(index int) int {
	if index > 0 {
		return index - 1
	} else {
		println("ERROR")
		return -1
	}
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	Term := rf.Term
	isleader := rf.IsLeader
	rf.mu.Unlock()
	return Term, isleader
}
func (rf *Raft) GetState2() (int, string) {
	rf.mu.Lock()
	Term := rf.Term
	var State string
	if rf.State == Follwer {
		State = "Follower"
	} else if rf.State == Cand {
		State = "Candidate"
	} else {
		State = "Leader"
	}
	rf.mu.Unlock()
	return Term, State
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
