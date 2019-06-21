package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
)

var data = map[string]int64{
	"a711355": 4500,
	// "a77d357": 100,
	// "a7f9365": 110,
	// "a77d397": 12,
}

func main() {
	path := "./data"
	rootHash := ""
	diskdb, err := ethdb.NewLDBDatabase(path, 0, 0)
	if err != nil {
		panic(err)
	}
	statedb := state.NewDatabase(diskdb)
	state, err := state.New(common.HexToHash(rootHash), statedb)
	if err != nil {
		panic(err)
	}

	for a, d := range data {
		addr := common.HexToAddress(a)
		state.CreateAccount(addr)
		state.SetBalance(addr, big.NewInt(d))
	}
	newHash, err := state.Commit(true)
	if err != nil {
		panic(err)
	}

	err = statedb.TrieDB().Commit(newHash, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Root hash %x\n", newHash)
}
