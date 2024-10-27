package gateway

import (
	"net"
	"tp1/internal/gateway/id_generator"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

// listenForNewClient listens for new clients and assigns them an unique client id
func (g *Gateway) listenForNewClient() error {
	return g.listenForConnections(ClientIdListener, g.assignClientId)
}

func (g *Gateway) assignClientId(c net.Conn) {
	g.IdGeneratorMu.Lock()
	clientId := g.IdGenerator.GetId()
	g.IdGeneratorMu.Unlock()

	err := ioutils.SendAll(c, id_generator.EncodeClientId(clientId))
	if err != nil {
		logs.Logger.Errorf("Error sending client id to client: %s", err)
	}
}
