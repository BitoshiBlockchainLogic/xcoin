// +build ignore

package main

import (
	"crypto/sha512"
	"fmt"
	"log"
)

func main() {

	coinPrivData := string("xGenesisKey" + "userPrivKey")
	h := sha512.New()
	h.Write([]byte(coinPrivData))
	coinPrivKey := h.Sum(nil)

	coinPubData := string("xGenesisKey" + string(coinPrivKey))
	h2 := sha512.New()
	h2.Write([]byte(coinPubData))
	coinPubKey := h2.Sum(nil)

	fmt.Printf("%x\n", coinPrivKey)
	log.Printf("Coin Private Key")
	fmt.Printf("%x\n", coinPubKey)
	log.Printf("Coin Public Key")

}
