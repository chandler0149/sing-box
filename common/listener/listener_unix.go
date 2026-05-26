package listener

import (
	"net"
	"os"
)

func (l *Listener) ListenUnix() (net.Listener, error) {
	_ = os.Remove(l.listenOptions.ListenPath)
	unixListener, err := ListenNetworkNamespace[net.Listener](l.listenOptions.NetNs, func() (net.Listener, error) {
		return net.Listen("unix", l.listenOptions.ListenPath)
	})
	if err != nil {
		return nil, err
	}
	l.logger.Info("unix server started at ", l.listenOptions.ListenPath)
	l.tcpListener = unixListener
	return unixListener, nil
}
