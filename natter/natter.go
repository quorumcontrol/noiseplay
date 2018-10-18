package main

import (
	"encoding/base64"
	"flag"

	"github.com/perlin-network/noise/crypto/ed25519"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/network"
	"github.com/perlin-network/noise/network/backoff"
	"github.com/perlin-network/noise/network/discovery"
	"github.com/quorumcontrol/noiseplay/natter/messages"
	"github.com/quorumcontrol/noiseplay/natter/plugin"
)

func main() {
	destination := flag.String("s", "", "is this the server?")

	flag.Parse()

	protocol := "tcp"
	host := "0.0.0.0"
	port := 10000
	keys := ed25519.RandomKeyPair()

	builder := network.NewBuilder()
	builder.SetKeys(keys)

	// Register NAT traversal plugin.
	// nat.RegisterPlugin(builder)

	// Register the reconnection plugin
	builder.AddPlugin(new(backoff.Plugin))

	// Register peer discovery plugin.
	builder.AddPlugin(new(discovery.Plugin))

	builder.AddPlugin(new(plugin.PingPlugin))

	if *destination == "" {
		builder.SetAddress(network.FormatAddress(protocol, host, uint16(port)))

		node, err := builder.Build()
		if err != nil {
			log.Fatal().Err(err).Msg("")
			panic(err)
		}
		log.Info().Msgf("server address: %s, id: %s", node.Address, base64.URLEncoding.EncodeToString(node.ID.Id))
		node.Listen()
	} else {
		builder.SetAddress(network.FormatAddress(protocol, host, uint16(port+1)))

		node, err := builder.Build()
		if err != nil {
			log.Fatal().Err(err).Msg("")
			panic(err)
		}
		log.Info().Msgf("client address: %s, id: %s", node.Address, node.ID.Id)
		go node.Listen()
		node.BlockUntilListening()
		serverAddress := network.FormatAddress(protocol, "172.21.0.2", uint16(port))
		log.Info().Msgf("client bootstrapping: %s", serverAddress)
		node.Bootstrap(serverAddress)
		log.Info().Msgf("client broadcasting")
		node.BroadcastByAddresses(&messages.NatPing{Payload: "hi"}, serverAddress)
		select {}
	}

}
