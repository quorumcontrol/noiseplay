package proxy

import (
	"github.com/perlin-network/noise/dht"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/network"
	"github.com/perlin-network/noise/network/discovery"
	"github.com/perlin-network/noise/peer"
	"github.com/pkg/errors"
	"github.com/quorumcontrol/noiseplay/p2p/messages"
)

// ProxyPlugin buffers all messages into a mailbox for this test.
type ProxyPlugin struct {
	*network.Plugin
	Mailbox chan *messages.ProxyMessage
}

func (n *ProxyPlugin) Startup(net *network.Network) {
	// Create mailbox.
	n.Mailbox = make(chan *messages.ProxyMessage, 1)
}

// Handle implements the network interface callback
func (n *ProxyPlugin) Receive(ctx *network.PluginContext) error {
	// Handle the proxy message.
	switch msg := ctx.Message().(type) {
	case *messages.ProxyMessage:
		// n.Mailbox <- msg

		log.Info().Msgf("%s Message received: %v", ctx.Network().Address, msg)
		//fmt.Fprintf(os.Stderr, "Node %d received a message from node %d.\n", ids[ctx.Network().Address], ids[ctx.Sender().Address])

		if err := n.ProxyBroadcast(ctx.Network(), ctx.Sender(), msg); err != nil {
			panic(err)
		}
	}
	return nil
}

func findPeer(routes *dht.RoutingTable, target peer.ID) *peer.ID {
	bucketID := target.XorID(routes.Self()).PrefixLen()
	bucket := routes.Bucket(bucketID)

	for e := bucket.Front(); e != nil; e = e.Next() {
		if e.Value.(peer.ID).Equals(target) {
			p := e.Value.(peer.ID)
			return &p
		}
	}
	return nil
}

// ProxyBroadcast proxies a message until it reaches a target ID destination.
func (n *ProxyPlugin) ProxyBroadcast(node *network.Network, sender peer.ID, msg *messages.ProxyMessage) error {
	log.Info().Msgf("%s - ProxyBroadcast", node.Address)
	targetID := peer.ID{
		Id: msg.Destination,
	}

	// Check if we are the target.
	if node.ID.Equals(targetID) {
		log.Info().Msgf("%s received message for me! %s", node.Address, msg.Payload)
		return nil
	}

	plugin, registered := node.Plugin(discovery.PluginID)
	if !registered {
		return errors.New("discovery plugin not registered")
	}

	routes := plugin.(*discovery.Plugin).Routes

	// If the target is in our routing table, directly proxy the message to them.
	if peer := findPeer(routes, targetID); peer != nil {
		log.Info().Msgf("%s found peer, sending directly", node.Address)
		node.BroadcastByIDs(msg, *peer)
		return nil
	}

	// Find the 2 closest peers from a nodes point of view (might include us).
	closestPeers := routes.FindClosestPeers(targetID, 2)
	log.Info().Msgf("%s found closest peers (before edit): %v", node.Address, closestPeers)

	// Remove sender and self from the list.
	for i, id := range closestPeers {
		log.Info().Msgf("%s - id: %s, node id: %s", node.Address, id.Address, node.ID.Address)
		if id.Equals(sender) || id.Equals(node.ID) {
			closestPeers = append(closestPeers[:i], closestPeers[i+1:]...)
			break
		}
	}

	// Seems we have ran out of peers to attempt to propagate to.
	if len(closestPeers) == 0 {
		return errors.Errorf("could not found route from node %s to node %s", node.ID, targetID)
	}
	log.Info().Msgf("%s found closest peers (after edit): %v", node.Address, closestPeers)

	// Propagate message to the closest peer.
	log.Info().Msgf("%s sending to: %s", node.Address, closestPeers[0].Address)
	node.BroadcastByIDs(msg, closestPeers[0])
	return nil
}
