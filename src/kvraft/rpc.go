package kvraft

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongLeader = "ErrWrongLeader"
)

const (
	Putt    = "Put"
	Appendd = "Append"
	Gett    = "Get"
)

type Err string

// Put or Append
type PutAppendArgs struct {
	Key    string
	Value  string
	Op     string // "Put" or "Append"
	Client int64
	Id     int64
}

type PutAppendReply struct {
	Err Err
}

type GetArgs struct {
	Key    string
	Client int64
	Id     int64
}

type GetReply struct {
	Err   Err
	Value string
}
