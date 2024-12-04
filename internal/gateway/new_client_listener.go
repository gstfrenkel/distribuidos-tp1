package gateway

import (
	"net"
	"tp1/pkg/logs"
	"tp1/pkg/utils/id_generator"
	"tp1/pkg/utils/io"
)

// listenForNewClient listens for new clients and assigns them an unique client id
func (g *Gateway) listenForNewClient() error {
	return g.listenForConnections(ClientIdListener, g.assignClientId)
}

func (g *Gateway) assignClientId(c net.Conn) {
	g.IdGeneratorMu.Lock()
	clientId := g.IdGenerator.GetId()
	g.IdGeneratorMu.Unlock()

	err := io.SendAll(c, id_generator.EncodeClientId(clientId))
	if err != nil {
		logs.Logger.Errorf("Error sending client id to client: %s", err)
	}
}
