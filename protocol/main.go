package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"trie-example/protocol/node"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
	yaml "gopkg.in/yaml.v2"
)

const (
	eth62 = 62
	eth63 = 63
)

var configPath *string

// setup logging
func init() {
	// ensure good log formats for terminal
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	loglevel := log.LvlTrace
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}
func main() {
	var configPath = "./config.yaml"
	var config node.Config
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Error("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	fmt.Print(config)

	privkey_one, err := crypto.HexToECDSA(config.Node.PrivateKey)

	if err != nil {
		fmt.Println("Generate private key failed", "err", err)
	}

	genesisHash := common.HexToHash(config.Node.GenesisHash)
	server := newServer(privkey_one, "phukq", "0.1", config.Node.Port, config.Node.NetworkId, genesisHash, genesisHash)

	server.Start()
	nodeinfo := server.NodeInfo()
	log.Info("server started ", "info", nodeinfo.Enode)

	// Connect to peer in the config file
	if config.Peer.Id != "" {
		url := fmt.Sprintf("enode://%s@%s:%d", config.Peer.Id, config.Peer.Ip, config.Peer.Port)
		fmt.Println("Connecting to ", url)
		node, err := enode.ParseV4(url)
		if err != nil {
			fmt.Errorf("invalid enode: %v", err)
		}
		server.AddPeer(node)
		fmt.Println(node)
	}

	// Keep running
	select {}
}
func newServer(privkey *ecdsa.PrivateKey, name string, version string, port int, networkId uint64, headHash, genesisHash common.Hash) *p2p.Server {
	pm, err := node.NewProtocolManager(networkId, headHash, genesisHash)
	if err != nil {
		panic(err)
	}
	// we need to explicitly allow at least one peer, otherwise the connection attempt will be refused
	cfg := p2p.Config{
		PrivateKey:      privkey,
		Name:            common.MakeName(name, version),
		MaxPeers:        1,
		Protocols:       pm.SubProtocols,
		EnableMsgEvents: true,
		NAT:             nat.Any(),
	}
	if port > 0 {
		cfg.ListenAddr = fmt.Sprintf(":%d", port)
	}
	srv := &p2p.Server{
		Config: cfg,
	}
	pm.Start()

	return srv
}
