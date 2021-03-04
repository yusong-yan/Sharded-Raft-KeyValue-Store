package kvraft

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"../labgob"
	"../labrpc"
	"../raft"
)

const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	OpTask string
	Key    string
	Value  string
	Client int64
	Id     int64
}

type KVServer struct {
	mu           sync.Mutex
	me           int
	rf           *raft.Raft
	applyCh      chan raft.ApplyMsg
	dead         int32 // set by Kill()
	maxraftstate int   // snapshot if log grows this big
	// Your definitions here.
	storage map[string]string
	idTable map[int64]map[int64]Op
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	if kv.isLeader() {
		op := Op{Gett, args.Key, "", args.Client, args.Id}
		_, _, Success := kv.rf.Start(op)
		if Success {
			for i := 0; i < 10; i++ {
				if kv.existIdWithLock(args.Id, args.Client) {
					if kv.existKeyWithLock(args.Key) {
						kv.mu.Lock()
						reply.Value = kv.storage[args.Key]
						kv.mu.Unlock()
					} else {
						reply.Value = ""
					}
					reply.Err = OK
					return
				}
				time.Sleep(20 * time.Millisecond)
			}
			reply.Err = ErrWrongLeader
			return
		} else {
			reply.Err = ErrWrongLeader
			return
		}
	} else {
		reply.Err = ErrWrongLeader
		return
	}
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	if kv.isLeader() {
		op := Op{args.Op, args.Key, args.Value, args.Client, args.Id}
		_, _, Success := kv.rf.Start(op)
		if Success {
			for i := 0; i < 10; i++ {
				if kv.existIdWithLock(args.Id, args.Client) {
					if kv.existKeyWithLock(args.Key) {
						reply.Err = OK
						return
					} else {
						fmt.Println("err")
						return
					}
				}
				time.Sleep(20 * time.Millisecond)
			}
			reply.Err = ErrWrongLeader
			return
		} else {
			reply.Err = ErrWrongLeader
			return
		}
	} else {
		reply.Err = ErrWrongLeader
		return
	}
}

func (kv *KVServer) isLeader() bool {
	_, isLeader := kv.rf.GetState()
	return isLeader
}

func (kv *KVServer) listenApplyCh() {
	for !kv.killed() {
		applyMessage := <-kv.applyCh
		var curOp Op
		curOp = applyMessage.Command.(Op)
		kv.mu.Lock()
		if !kv.existId(curOp.Id, curOp.Client) {
			if curOp.OpTask == Appendd && kv.existKey(curOp.Key) {
				kv.storage[curOp.Key] += curOp.Value
			} else if curOp.OpTask == Putt || curOp.OpTask == Appendd {
				kv.storage[curOp.Key] = curOp.Value
			}
		}
		_, exist := kv.idTable[curOp.Client]
		if !exist {
			kv.idTable[curOp.Client] = map[int64]Op{}
		}
		kv.idTable[curOp.Client][curOp.Id] = curOp

		kv.mu.Unlock()
	}
}

func (kv *KVServer) existId(id int64, client int64) bool {
	_, exist := kv.idTable[client]
	if exist == true {
		_, exist = kv.idTable[client][id]
		if exist == true {
			return true
		}
	}
	return false
}

func (kv *KVServer) existIdWithLock(id int64, client int64) bool {
	kv.mu.Lock()
	e := kv.existId(id, client)
	kv.mu.Unlock()
	return e
}

func (kv *KVServer) existKey(id string) bool {
	_, exist := kv.storage[id]
	return exist
}

func (kv *KVServer) existKeyWithLock(id string) bool {
	kv.mu.Lock()
	e := kv.existKey(id)
	kv.mu.Unlock()
	return e
}

func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	// Your code here, if desired.
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	labgob.Register(Op{})
	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.applyCh = make(chan raft.ApplyMsg, 1)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.storage = map[string]string{}
	kv.idTable = map[int64]map[int64]Op{}
	go kv.listenApplyCh()
	return kv
}
