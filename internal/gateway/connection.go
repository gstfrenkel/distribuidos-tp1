package gateway

import (
	"net"
	"tp1/pkg/logs"
)

const LenFieldSize = 4
const TransportProtocol = "tcp"

func (g *Gateway) createGatewaySockets() error {
	gamesListener, err := g.newListener("gateway.games-address")
	if err != nil {
		return err
	}
	reviewsListener, err := g.newListener("gateway.reviews-address")
	if err != nil {
		return err
	}

	resultsListener, err := g.newListener("gateway.results-address")
	if err != nil {
		return err
	}

	clientIdListener, err := g.newListener("gateway.client-id-address")
	if err != nil {
		return err
	}

	g.Listeners[GamesListener] = gamesListener
	g.Listeners[ReviewsListener] = reviewsListener
	g.Listeners[ResultsListener] = resultsListener
	g.Listeners[ClientIdListener] = clientIdListener

	logs.Logger.Infof("Gateway listening client id on %s", clientIdListener.Addr().String())
	logs.Logger.Infof("Gateway listening games on %s", gamesListener.Addr().String())
	logs.Logger.Infof("Gateway listening reviews on %s", reviewsListener.Addr().String())
	logs.Logger.Infof("Gateway listening results on %s", resultsListener.Addr().String())
	return nil
}

func (g *Gateway) newListener(configKey string) (net.Listener, error) {
	addr := g.Config.String(configKey, "")
	return net.Listen(TransportProtocol, addr)
}

func (g *Gateway) listenForConnections(listener int, handleConnection func(net.Conn)) error {
	logs.Logger.Infof("Waiting for new client connections, listener %d...", listener)
	for {
		g.finishedMu.Lock()
		if g.finished {
			g.finishedMu.Unlock()
			break
		}
		g.finishedMu.Unlock()
		c, err := g.Listeners[listener].Accept()
		logs.Logger.Infof("Successfully established new connection! Listener %d...", listener)
		if err != nil {
			return err
		}
		go handleConnection(c)
	}
	return nil
}
