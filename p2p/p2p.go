package p2p

import (
	"time"

	"github.com/perlin-network/noise/crypto/ed25519"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/network"
	"github.com/perlin-network/noise/network/backoff"
	"github.com/perlin-network/noise/network/discovery"
	"github.com/quorumcontrol/noiseplay/p2p/messages"
	"github.com/quorumcontrol/noiseplay/p2p/proxy"
)

func RunP2P() error {
	protocol := "kcp"
	host := "127.0.0.1"

	numNodes := 20
	sender := 0
	target := numNodes - 1

	nodes := make([]*network.Network, numNodes)
	port := 10000

	for i := 0; i < numNodes; i++ {

		keys := ed25519.RandomKeyPair()

		builder := network.NewBuilder()
		builder.SetKeys(keys)
		builder.SetAddress(network.FormatAddress(protocol, host, uint16(port+i)))

		// Register NAT traversal plugin.
		// nat.RegisterPlugin(builder)

		// Register the reconnection plugin
		builder.AddPlugin(new(backoff.Plugin))

		// Register peer discovery plugin.
		builder.AddPlugin(&discovery.Plugin{
			DisablePong: true,
		})

		// Register proxy plugin
		builder.AddPlugin(new(proxy.ProxyPlugin))

		node, err := builder.Build()
		if err != nil {
			log.Fatal().Err(err).Msg("")
			return err
		}

		log.Info().Msgf("node address: %s, id: %s", node.Address, node.ID.Id)
		nodes[i] = node
		go node.Listen()
	}

	for _, node := range nodes {
		node.BlockUntilListening()
	}

	for i := 0; i < numNodes; i++ {
		var peerList []string
		if i > 0 {
			peerList = append(peerList, nodes[i-1].Address)
		}
		if i < numNodes-1 {
			peerList = append(peerList, nodes[i+1].Address)
		}

		// Bootstrap nodes to their assignd peers.
		nodes[i].Bootstrap(peerList...)
	}

	msg := &messages.ProxyMessage{
		Payload:     []byte("hi"),
		Destination: nodes[target].ID.Id,
		Source:      nodes[sender].ID.Id,
	}

	nodes[sender].BroadcastByIDs(msg, nodes[sender+1].ID)

	<-time.After(20 * time.Second)

	return nil
}
