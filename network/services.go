package network

import (
	"time"
	"github.com/bazo-blockchain/bazo-client/util"
)

//Single goroutine that makes sure the system is well connected.
func checkHealthService() {
	for {
		time.Sleep(util.HEALTH_CHECK_INTERVAL * time.Second)

		//Stop trying to connect after #numberOfRetry attempts.
		numberOfRetry := 0

		if !peers.contains(util.Config.BootstrapIpport) {
			p, err := initiateNewClientConnection(util.Config.BootstrapIpport)
			if p == nil || err != nil {
				logger.Printf("%v\n", err)
			} else {
				go minerConn(p)
			}
		}

		//Periodically check if we are well-connected
		if len(peers.minerConns) >= util.MIN_MINERS {
			continue
		}

		//The only goto in the code (I promise), but best solution here IMHO.
	RETRY:
		select {
		//iplistChan gets filled with every incoming neighborRes, they're consumed here.
		case ipaddr := <-iplistChan:
			p, err := initiateNewClientConnection(ipaddr)
			if err != nil {
				logger.Printf("%v\n", err)
			}

			if p == nil || err != nil {
				if numberOfRetry < 3 {
					numberOfRetry++
					goto RETRY
				} else {
					break
				}
			}

			go minerConn(p)
			break
		default:
			//In case we don't have any ip addresses in the channel left, make a request to the network.
			neighborReq()
			break
		}
	}
}

func peerService() {
	for {
		select {
		case p := <-register:
			peers.add(p)
		case p := <-disconnect:
			peers.delete(p)
			close(p.ch)
		}
	}
}
