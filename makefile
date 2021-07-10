#######Inorder to run client and server, modify src/runner/srunner/main/servers.txt
#######to adjust number of machine, and also it's serveraddress with it's port

#test raft protocal and storage system
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

# Run client or server.
ip:
	cd src/runner/ip/main && go run ip.go
client:
	cd src/runner/crunner/main && go run cRun.go
server:
	cd src/runner/srunner/main && go run sRun.go
cleanPersist:
	cd src/runner/srunner/main && rm *.yys
# Usually you should run 3 or more servers and as many clients as you want

# for server, you should give a port number
# because the number of servers machine should be the same all the time
# checkout src/runner/srunner/main/servers.txt to see the port numbers, and choose one of it
