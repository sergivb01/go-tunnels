package mcserver

import "sync/atomic"

type AtomBool bool

type RemoteConn struct {
	Address string
	Port    int
	found   uintptr
}

func (r *RemoteConn) SetFound() {
	atomic.StoreUintptr(&r.found, 1)
}
