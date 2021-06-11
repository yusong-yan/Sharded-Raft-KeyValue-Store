raftTestAll: 
	cd test/raft && go test -run 2
raftTest2A: 
	cd test/raft && go test -run 2A
raftTest2B: 
	cd test/raft && go test -run 2B
raftTest2C: 
	cd test/raft && go test -run 2C
kvTestAll: 
	cd test/kvraft && go test -run 3
kvTest3A: 
	cd test/kvraft && go test -run 3A
kvTest3B: 
	cd test/kvraft && go test -run 3B
client:
	cd src/runner/crunner/main && go run crunner.go
server:
	cd src/runner/srunner/main && go run srunner.go