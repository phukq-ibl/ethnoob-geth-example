package main

import (
	"fmt"

	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/ethereum/go-ethereum/ethdb"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var data = map[string]string{
	"do":    "verb",
	"dog":   "puppy",
	"doge":  "coin",
	"horse": "stallion",
}

func main() {
	path := "./data"
	rootHash := ""

	// rootHash = "5c00467a8a08d5d233e83d3ef757daaaf76bad07803bd96492c36455a93af3de"
	// writeDataToTrie(rootHash, path)
	writeDataMap(rootHash, path)
	// getTrie(rootHash, path, ab)

	iterateDb(path)

}

func writeDataMap(rootHash string, path string) {
	// Create new DB instance
	diskdb, err := ethdb.NewLDBDatabase(path, 0, 0)
	if err != nil {
		panic(err)
	}

	// Create a triedb
	triedb := trie.NewDatabase(diskdb)

	// Create a trie with the root hash is nil
	trie, err := trie.New(common.HexToHash(rootHash), triedb)

	// Insert data to the trie
	for k, v := range data {
		// fmt.Println(v, k)
		err = trie.TryUpdate([]byte(k), []byte(v))
		if err != nil {
			panic(err)
		}
	}
	// Commit trie
	newRoothash, err := trie.Commit(nil)

	if err != nil {
		panic(err)
	}
	fmt.Println("new root ", (newRoothash.Hex()))

	// Write data to disk
	triedb.Commit(newRoothash, true)
	diskdb.Close()
}

func getTrie(rootHash, path string, key []byte) {
	diskdb, err := ethdb.NewLDBDatabase(path, 0, 0)
	if err != nil {
		panic(err)
	}
	triedb := trie.NewDatabase(diskdb)

	trie, err := trie.New(common.HexToHash(rootHash), triedb)
	val := trie.Get(key)
	fmt.Println("rs ", val)
}

func iterateDb(file string) {
	handles := 16
	cache := 16
	db, err := leveldb.OpenFile(file, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		value := iter.Value()
		// elems, _, err := rlp.SplitList(value)
		// if err != nil {
		// 	fmt.Errorf("decode error: %v", err)
		// }
		// count, err := rlp.CountValues(elems)
		// if err != nil {
		// 	fmt.Println("Err ", err)
		// }
		// fmt.Println("count ", count)
		fmt.Printf("%x=>%x\n", key, value)
	}
	iter.Release()
	err = db.Close()
	if err != nil {
		panic(err)
	}
}
