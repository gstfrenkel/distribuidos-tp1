package gateway

import (
	"net"
	"tp1/pkg/logs"
)

const (
	LenFieldSize      = 4
	gamesAddrKey      = "gateway.games-address"
	TransportProtocol = "tcp"
	reviewsAddrKey    = "gateway.reviews-address"
	resultsAddrKey    = "gateway.results-address"
	clientIdAddrKey   = "gateway.client-id-address"
)

func (g *Gateway) createGatewaySockets() error {
	gamesListener, err := g.newListener(gamesAddrKey)
	if err != nil {
		return err
	}
	reviewsListener, err := g.newListener(reviewsAddrKey)
	if err != nil {
		return err
	}

	resultsListener, err := g.newListener(resultsAddrKey)
	if err != nil {
		return err
	}

	clientIdListener, err := g.newListener(clientIdAddrKey)
	if err != nil {
		return err
	}

	g.Listeners[GamesListener] = gamesListener
	g.Listeners[ReviewsListener] = reviewsListener
	g.Listeners[ResultsListener] = resultsListener
	g.Listeners[ClientIdListener] = clientIdListener

	g.logListeners("games", gamesListener)
	g.logListeners("reviews", reviewsListener)
	g.logListeners("client id", clientIdListener)
	g.logListeners("results", resultsListener)
	return nil
}

func (g *Gateway) logListeners(listenerName string, listener net.Listener) {
	logs.Logger.Infof("Gateway listening %s on %s", listenerName, listener.Addr().String())
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
