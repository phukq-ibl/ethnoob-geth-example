package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {

	var tx1 = "0xf8608080830493e094ae78a9031d8b94aebc5a0bf1c4224411509219a780801ba018c6b96e0fe47bb59f5da7eade99df4d2fb1d9630cba6e690711587355e3ba71a07833a77fc9788ad2e6063289cbb9012e1083906124057770c40a0e89123c3f30"
	tx1 = "0xf8608010830f424094ae78a9031d8b94aebc5a0bf1c4224411509219a701801ba08885255cc6d83d338ed663a5804a152201137fe6074224b15f232f1d11c60957a07126b0a4ac2c667a3bfb1d9540ebed248c677675381f867219320696e5704f02" //6
	txOne := parseTx(tx1)
	signer := types.NewEIP155Signer(big.NewInt(1))
	msg, _ := txOne.AsMessage(signer)
	fmt.Println(msg.From().Hex())
}

func parseTx(hex string) *types.Transaction {
	tx := new(types.Transaction)
	encodedTx, parseError := hexutil.Decode(hex)
	if parseError != nil {
		msg := fmt.Sprintf("File error: %v\n", parseError)
		panic(msg)
	}

	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		fmt.Println("Can not parse tx")
		panic(err)
	}

	return tx
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
		fmt.Println(key, "=>", value)
	}
	iter.Release()
}
