package plugin

import (
	"time"

	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/network"
	"github.com/perlin-network/noise/peer"
	"github.com/quorumcontrol/noiseplay/natter/messages"
)

// ProxyPlugin buffers all messages into a mailbox for this test.
type PingPlugin struct {
	*network.Plugin
}

func (n *PingPlugin) Startup(net *network.Network) {
	// nil
}

// Handle implements the network interface callback
func (n *PingPlugin) Receive(ctx *network.PluginContext) error {
	// Handle the proxy message.
	switch msg := ctx.Message().(type) {
	case *messages.NatPing:
		log.Info().Msgf("%s Message received: %v", ctx.Network().Address, msg)
		//fmt.Fprintf(os.Stderr, "Node %d received a message from node %d.\n", ids[ctx.Network().Address], ids[ctx.Sender().Address])

		go func(node *network.Network, sender peer.ID, msg *messages.NatPing) {
			time.Sleep(1 * time.Second)
			log.Info().Msgf("%s responding with pong: %v", ctx.Network().Address, msg)
			node.BroadcastByIDs(&messages.NatPong{Payload: msg.Payload}, sender)
		}(ctx.Network(), ctx.Sender(), msg)
	case *messages.NatPong:
		log.Info().Msgf("%s message received: %s", ctx.Network().Address, msg.Payload)
	}
	return nil
}
