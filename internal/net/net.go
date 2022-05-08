package net

import (
	"sync"

	"github.com/mickep76/netlink"
)

const (
	sysNetPath = "/sys/class/net"
)

var (
	sharedNetHandler *netHandler
)

type netHandler struct {
	cond *sync.Cond

	ifaces []netlink.Interface
}

func getNetHandler() (*netHandler, error) {
	if sharedNetHandler != nil {
		return sharedNetHandler, nil
	}

	mu := new(sync.Mutex)

	sharedNetHandler = &netHandler{
		cond: sync.NewCond(mu),
	}
	sharedNetHandler.cond.L.Lock()

	if err := sharedNetHandler.updateIFaces(); err != nil {
		return nil, err
	}

	return sharedNetHandler, nil
}

func (n *netHandler) updateIFaces() error {
	ifs, err := netlink.Interfaces()
	if err != nil {
		return err
	}

	n.ifaces = ifs

	return nil
}
