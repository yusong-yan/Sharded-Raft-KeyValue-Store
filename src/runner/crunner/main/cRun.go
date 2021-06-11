package main

// This file is for running the server
import (
	"Sharded-RaftKV/src/kvraft"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"
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
	servers := readfile("../../srunner/main/servers.txt")
	kvraft.MakeClerkRun(servers)
}
