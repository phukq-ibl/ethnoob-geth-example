package node

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	eth "github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
)

const (
	forceSyncCycle      = 3 * time.Second // Time interval to force syncs, even if few peers are available
	minDesiredPeerCount = 5               // Amount of peers desired to start syncing
)

var bodiesQueue = make(map[string][]*types.Header)

const PROTOCOL_VERSION = 63

type ProtocolManager struct {
	networkID    uint64
	SubProtocols []p2p.Protocol
	newPeerCh    chan *peer

	peers                 *peerSet
	headHash, genesisHash common.Hash
}

func NewProtocolManager(networkID uint64, headHash, genesisHash common.Hash) (*ProtocolManager, error) {
	manager := &ProtocolManager{
		networkID:   networkID,
		newPeerCh:   make(chan *peer),
		peers:       newPeerSet(),
		headHash:    headHash,
		genesisHash: genesisHash,
	}

	proto := p2p.Protocol{
		Name:    eth.ProtocolName,
		Version: PROTOCOL_VERSION,
		Length:  17,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			peer := newPeer(PROTOCOL_VERSION, p, rw)

			select {
			case manager.newPeerCh <- peer:
				return manager.handle(peer)
			}
			return nil
		},
	}

	manager.SubProtocols = []p2p.Protocol{proto}

	return manager, nil
}

func (pm *ProtocolManager) Start() {
	log.Debug("====> Start manager protocol")

	go pm.syncer()
}

func (pm *ProtocolManager) handle(p *peer) error {
	log.Debug("Handle peer ", p.id)
	var (
		td          = big.NewInt(0)
		headHash    = pm.headHash
		genesisHash = pm.genesisHash
	)

	if err := p.Handshake(pm.networkID, td, headHash, genesisHash); err != nil {
		p.Log().Error("======> Ethereum handshake failed", "err", err)
		return err
	}

	// Add peer to peers list
	if err := pm.peers.Register(p); err != nil {
		p.Disconnect(p2p.DiscTooManyPeers)
		p.Log().Error("Ethereum peer registration failed", "err", err)
		return err
	}

	// Handle peer
	for {
		if err := pm.handleMsg(p); err != nil {
			return err
		}
	}
	return nil
}

func (pm *ProtocolManager) handleMsg(p *peer) error {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}

	switch {
	case msg.Code == eth.BlockHeadersMsg:
		var headers []*types.Header
		if err := msg.Decode(&headers); err != nil {
			return err
		}

		hashes := []common.Hash{}
		log.Info("==> Receive a  header ", "len", len(headers))
		for _, h := range headers {
			hashes = append(hashes, h.Hash())
			log.Info("header ", "number ", h.Number, "hash", h.Hash().Hex())
		}
		bodiesQueue[p.ID().String()] = headers
		err := RequestBodiesByHash(p.rw, hashes)
		if err != nil {
			log.Error("err", err)
			return err
		}
	case msg.Code == eth.BlockBodiesMsg:
		var request blockBodiesData
		if err := msg.Decode(&request); err != nil {
			return err
		}

		blocks := RestructBodies(bodiesQueue[p.ID().String()], request)

		log.Info("==> Receive block ", "len", len(blocks))
		hashes := []common.Hash{}
		for _, b := range blocks {
			hashes = append(hashes, b.Hash())
			log.Info("block ", "number ", b.Number(), "nonce ", b.Nonce(), "uncle", len(b.Uncles()), "tx", len(b.Transactions()))
		}

		// Send request receipt
		go RequestReceiptByHash(p.rw, hashes)
	case msg.Code == eth.ReceiptsMsg:
		var receipts [][]*types.Receipt
		if err := msg.Decode(&receipts); err != nil {
			return err
		}
		log.Info("=======> Receive receipts ", "len", len(receipts))
		RestructReceipts(bodiesQueue[p.ID().String()], receipts)
		for _, r := range receipts {
			for _, l := range r {
				log.Info("Receipt ", "hash", l.TxHash, "gas used", l.GasUsed)
			}
		}
	case msg.Code == eth.TxMsg:
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(1, "msg %v: %v", msg, err)
		}
		log.Info("========> Receive txs ", "len", len(txs))
		for i, tx := range txs {
			if tx == nil {
				return errResp(1, "transaction %d is nil", i)
			}
			log.Info("Tx", "index", i, "hash", tx.Hash())
		}
	default:
		log.Error("No handle yet ", "msg code", msg.Code)
	}
	return nil
}

func (pm *ProtocolManager) syncer() {
	forceSync := time.NewTicker(forceSyncCycle)
	defer forceSync.Stop()
	for {
		select {
		case <-pm.newPeerCh:
			// Make sure we have peers to select from, then sync
			if pm.peers.Len() < minDesiredPeerCount {
				break
			}
			go pm.synchronise(pm.peers.BestPeer())

		case <-forceSync.C:
			// Force a sync even if not enough peers are present
			log.Debug("========> Force sync")
			pm.synchronise(pm.peers.BestPeer())

			// case <-pm.noMorePeers:
			// 	return
		}
	}
}
func (pm *ProtocolManager) synchronise(peer *peer) {
	if peer == nil {
		log.Debug("Best peer is nil")
		return
	}

	err := SendGetBlockHeader(peer.rw, pm.genesisHash, 10, 0)

	if err != nil {
		log.Error("Get block header", "error", err)
	}

}

func SendGetBlockHeader(rw p2p.MsgReadWriter, origin common.Hash, amount, skip int) error {
	log.Debug("SendGetBlockHeader", "origin", origin, "amount", amount)
	return p2p.Send(
		rw,
		eth.GetBlockHeadersMsg,
		&getBlockHeadersData{Origin: hashOrNumber{Hash: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: false},
	)

}

func RequestBodiesByHash(rw p2p.MsgWriter, hashes []common.Hash) error {
	log.Debug("RequestBodiesByHash ", "len", len(hashes))
	return p2p.Send(rw, eth.GetBlockBodiesMsg, hashes)
}

func RequestReceiptByHash(rw p2p.MsgWriter, hashes []common.Hash) error {
	log.Debug("Request receipt ", "len", len(hashes))
	return p2p.Send(rw, eth.GetReceiptsMsg, hashes)
}

func RestructBodies(h []*types.Header, request blockBodiesData) []*types.Block {
	blocks := make([]*types.Block, 0)
	for i, r := range request {
		txHash := types.DeriveSha(types.Transactions(r.Transactions))
		if txHash != h[i].TxHash {
			panic("Not match tx hash")
		}
		block := types.NewBlockWithHeader(h[i]).WithBody(r.Transactions, r.Uncles)
		blocks = append(blocks, block)
	}
	return blocks
}

func RestructReceipts(h []*types.Header, receipts [][]*types.Receipt) {
	for i, r := range receipts {
		receiptHash := types.DeriveSha(types.Receipts(r))
		if receiptHash != h[i].ReceiptHash {
			panic("Not match tx hash")
		}
	}
}
