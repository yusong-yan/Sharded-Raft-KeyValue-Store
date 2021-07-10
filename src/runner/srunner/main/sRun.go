package main

// This file is for running the server
import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"unicode"

	"Sharded-RaftKV/src/kvraft"
)

func readfile(fileName string) []string {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		log.Fatalf("cannot open " + fileName)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", file)
	}
	ff := func(r rune) bool { return unicode.IsSpace(r) }
	words := strings.FieldsFunc(string(content), ff)
	return words
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {
	ip := GetOutboundIP().String()
	servers := readfile("servers.txt")
	fmt.Println(servers)
	me := -1
	for k, v := range servers {
		if strings.Contains(v, ip) {
			me = k
			break
		}
	}
	if me == -1 {
		panic("Couldn't find port in the server.txt")
	}
	kvraft.StartKVServerRun(servers, me, -1)
}
