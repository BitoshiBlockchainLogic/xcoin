package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

// Coin represents each 'item' in the blockchain
type Coin struct {
	Index     int
	Timestamp string
	genKey    string
	Hash      string
	PrevHash  string
}

// Coinchain is a series of validated Coins
var Coinchain []Coin

// SHA256 hashing
func calculateHash(coin Coin) string {
	record := string(coin.Index) + coin.Timestamp + string(coin.genKey) + coin.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new coin using previous coin's hash
func generateCoin(oldCoin Coin, genKey string) (Coin, error) {

	var newCoin Coin

	t := time.Now()

	newCoin.Index = oldCoin.Index + 1
	newCoin.Timestamp = t.String()
	newCoin.genKey = genKey
	newCoin.PrevHash = oldCoin.Hash
	newCoin.Hash = calculateHash(newCoin)

	return newCoin, nil
}

// make sure coin is valid by checking index, and comparing the hash of the previous coin
func isCoinValid(newCoin, oldCoin Coin) bool {
	if oldCoin.Index+1 != newCoin.Index {
		return false
	}

	if oldCoin.Hash != newCoin.PrevHash {
		return false
	}

	if calculateHash(newCoin) != newCoin.Hash {
		return false
	}

	return true
}

// make sure the chain we're checking is longer than the current blockchain
func replaceChain(newCoins []Coin) {
	if len(newCoins) > len(Coinchain) {
		Coinchain = newCoins
	}
}

// bcServer handles incoming concurrent Coins
var bcServer chan []Coin

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bcServer = make(chan []Coin)

	// create genesis coin
	t := time.Now()
	genesisCoin := Coin{0, t.String(), "", "", ""}
	spew.Dump(genesisCoin)
	Coinchain = append(Coinchain, genesisCoin)

	// start TCP and serve TCP server
	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {

	io.WriteString(conn, "Enter a new Genesis Key:")

	scanner := bufio.NewScanner(conn)

	// take in genKey from stdin and add it to blockchain after conducting necessary validation
	go func() {

		for scanner.Scan() {
			var genKey string
			genKey, err := scanner.Text(), scanner.Err()

			if err != nil {
				log.Printf("%v error: %v", scanner.Text(), err)
				continue
			}

			newCoin, err := generateCoin(Coinchain[len(Coinchain)-1], genKey)
			if err != nil {
				log.Println(err)
				continue
			}
			if isCoinValid(newCoin, Coinchain[len(Coinchain)-1]) {
				newCoinchain := append(Coinchain, newCoin)
				replaceChain(newCoinchain)
			}

			bcServer <- Coinchain
			io.WriteString(conn, "\nEnter a new genKey:")
		}
	}()

	defer conn.Close()

	// simulate receiving broadcast
	go func() {
		for {
			time.Sleep(30 * time.Second)
			output, err := json.Marshal(Coinchain)
			if err != nil {
				log.Fatal(err)
			}
			io.WriteString(conn, string(output))
		}
	}()

	for _ = range bcServer {
		spew.Dump(Coinchain)
	}

}
