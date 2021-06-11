package main

import (
	"fmt"
	"io/ioutil"
	"log"
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
	ff := func(r rune) bool { return !unicode.IsNumber(r) && !unicode.IsLetter(r) }
	words := strings.FieldsFunc(string(content), ff)
	return words
}

func main() {
	fmt.Print("Please provide host port: ")
	var port string
	fmt.Scan(&port)
	servers := readfile("servers.txt")
	me := -1
	for k, v := range servers {
		if v == port {
			me = k
			break
		}
	}
	if me == -1 {
		panic("Couldn't find port in the server.txt")
	}
	kvraft.StartKVServerRun(servers, me, -1)
}
