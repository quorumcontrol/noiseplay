package p2p

import (
	"github.com/perlin-network/noise/crypto/ed25519"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/network"
	"github.com/perlin-network/noise/network/backoff"
	"github.com/perlin-network/noise/network/discovery"
	"github.com/perlin-network/noise/network/nat"
)

func RunP2P() error {
	port := uint16(network.GetRandomUnusedPort())
	protocol := "kcp"
	host := "0.0.0.0"

	keys := ed25519.RandomKeyPair()

	log.Info().Str("private_key", keys.PrivateKeyHex()).Msg("")
	log.Info().Str("public_key", keys.PublicKeyHex()).Msg("")

	builder := network.NewBuilder()
	builder.SetKeys(keys)
	builder.SetAddress(network.FormatAddress(protocol, host, port))

	// Register NAT traversal plugin.
	nat.RegisterPlugin(builder)

	// Register the reconnection plugin
	builder.AddPlugin(new(backoff.Plugin))

	// Register peer discovery plugin.
	builder.AddPlugin(new(discovery.Plugin))

	node, err := builder.Build()
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return err
	}

	node.Listen()

	return nil
}
