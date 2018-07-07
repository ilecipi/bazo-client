package network

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
