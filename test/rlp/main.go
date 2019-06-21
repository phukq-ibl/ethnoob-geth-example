package main

import "github.com/ethereum/go-ethereum/rlp"

func main() {
	data := "8080c8823162844242424280808080808080808080"
	rlp.Decode(data)
}
