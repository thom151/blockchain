package main

import (
	//	"bytes"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	//	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	//	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

type Message struct {
	BPM int
}

var Blockchain []Block

var bcServer chan []Block

func main() {
	log.Println("Hello World")

	err := godotenv.Load()
	if err != nil {
		log.Println("cannot find environment variables")
		return
	}

	bcServer = make(chan []Block)

	go func() {
		t := time.Now()
		genesisBlock := Block{
			Index:     0,
			Timestamp: t.String(),
			BPM:       0,
			Hash:      "",
			PrevHash:  "",
		}

		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()

	server, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Println("cannot listen to tcp server")
		return
	}

	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "BPM: ")

	scanner := bufio.NewScanner(conn)

	go func() {
		for scanner.Scan() { //as long as there are still token
			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("errorr converting %v to a number\n", scanner.Text())
				continue
			}

			newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], bpm)
			if err != nil {
				log.Printf("error generating new block: %v\n", err)
				continue
			}

			if isBlockValid(Blockchain[len(Blockchain)-1], newBlock) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)

			}

			bcServer <- Blockchain
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()

	go func() {
		for {
			time.Sleep(30 * time.Second)
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}

			io.WriteString(conn, string(output))
		}
	}()

	for _ = range bcServer {
		spew.Dump(Blockchain)
	}

}

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash

	h := sha256.New()

	h.Write([]byte(record))
	hashed := h.Sum(nil)

	return hex.EncodeToString(hashed)
}

func generateBlock(prevBlock Block, currBPM int) (Block, error) {
	var newBlock Block

	newBlock.PrevHash = prevBlock.Hash
	newBlock.Index = prevBlock.Index + 1

	newBlock.Timestamp = time.Now().UTC().String()

	newBlock.BPM = currBPM

	newBlock.Hash = calculateHash(newBlock)
	return newBlock, nil
}

func isBlockValid(oldBlock, newBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}
