package httpApiServer

import (
	"moonlighting/communityServiceTradingCenter/dataManager"
	"net"
	"net/http"
	"sync"
)

type Server struct {
	listenAddress        string
	netListener          net.Listener
	providerDataManager  *dataManager.Manager
	publishDataManager   *dataManager.Manager
	recommendDataManager *dataManager.Manager
	staticServePath      string
	stopSignal           chan int
	stopOnce             sync.Once
}

func NewHttpApiServer(listenAddress string, htmlServePath string, proDm *dataManager.Manager, pubDm *dataManager.Manager, recDm *dataManager.Manager) *Server {
	netListener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		panic(err)
	}
	res := &Server{
		listenAddress:        listenAddress,
		netListener:          netListener,
		providerDataManager:  proDm,
		publishDataManager:   pubDm,
		recommendDataManager: recDm,
		staticServePath:      htmlServePath,
		stopSignal:           make(chan int),
		stopOnce:             sync.Once{},
	}

	return res

}

func (p *Server) Start() {
	err := http.Serve(p.netListener, p.route())
	if err != nil {
		p.Stop()
		return
	}

	<-p.stopSignal
}

func (p *Server) Stop() {
	p.stopOnce.Do(func() {
		select {
		case <-p.stopSignal:
			{
				return
			}
		default:

		}
		_ = p.netListener.Close()
		close(p.stopSignal)
	})
}
