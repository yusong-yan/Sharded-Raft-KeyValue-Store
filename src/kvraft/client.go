package kvraft

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"Sharded-RaftKV/src/labrpc"
)

type Clerk struct {
	servers  []*labrpc.ClientEnd
	mu       sync.Mutex
	clientId int64

	Test       bool     //for run
	serversRun []string //for run

	serverNumber int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.serverNumber = len(servers)
	ck.clientId = time.Now().UnixNano()
	ck.Test = true
	return ck
}

func (ck *Clerk) Get(key string) string {
	ck.mu.Lock()
	defer ck.mu.Unlock()
	args := GetArgs{}
	args.Id = time.Now().UnixNano()
	args.Client = ck.clientId
	args.Key = key
	for {
		for i := 0; i < ck.serverNumber; i++ {
			reply := GetReply{}
			var ok bool
			if ck.Test {
				ok = ck.servers[i].Call("KVServer.Get", &args, &reply)
			} else {
				ok = call("KVServer.GetRun", ck.serversRun[i], &args, &reply)
			}
			if ok {
				if reply.Err == OK {
					return reply.Value
				}
			}
		}
		fmt.Println("Retrying")
		time.Sleep(50 * time.Microsecond)
	}
}

func (ck *Clerk) PutAppend(key string, value string, op string) {
	// You will have to modify this function.
	ck.mu.Lock()
	defer ck.mu.Unlock()
	args := PutAppendArgs{}
	args.Op = op
	args.Value = value
	args.Client = ck.clientId
	args.Key = key
	args.Id = time.Now().UnixNano()
	for {
		for i := 0; i < ck.serverNumber; i++ {
			reply := PutAppendReply{}
			var ok bool
			if ck.Test {
				ok = ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
			} else {
				ok = call("KVServer.PutAppendRun", ck.serversRun[i], &args, &reply)
			}

			if ok {
				if reply.Err == OK {
					return
				}
			}
			//time.Sleep(20 * time.Millisecond)
		}
		fmt.Println("Retrying")
		time.Sleep(50 * time.Microsecond)
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, Putt)
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, Appendd)
}
