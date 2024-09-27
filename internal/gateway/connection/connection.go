package connection

import (
	"net"
	"tp1/internal/gateway"
)

func CreateGatewaySocket(g *gateway.Gateway) error {
	conn, err := net.Listen("tcp", g.Config.String("gateway.address", ""))
	if err != nil {
		return err
	}

	g.Listener = conn
	return nil
}

func ListenForNewClients(g *gateway.Gateway) error {
	for {
		c, err := g.Listener.Accept()
		if err != nil {
			return err
		}
		go handleConnection(g, c)
	}
}

func handleConnection(g *gateway.Gateway, conn net.Conn) {

}

func sendAll(conn net.Conn, data []byte) error {
	total := len(data)
	for total > 0 {
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
		total -= n
	}
	return nil
}
