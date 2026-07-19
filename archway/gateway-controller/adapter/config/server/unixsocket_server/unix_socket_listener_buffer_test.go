package unixsocket_server

import "net"

const benchmarkUnixSocketBufferBytes = 1 << 20 // 1 MiB

type unixConnBufferedListener struct {
	net.Listener
	readBufferBytes  int
	writeBufferBytes int
}

func newUnixConnBufferedListener(socketPath string, readBufferBytes, writeBufferBytes int) (net.Listener, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	if readBufferBytes <= 0 && writeBufferBytes <= 0 {
		return listener, nil
	}

	return &unixConnBufferedListener{
		Listener:         listener,
		readBufferBytes:  readBufferBytes,
		writeBufferBytes: writeBufferBytes,
	}, nil
}

func (l *unixConnBufferedListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return conn, nil
	}

	if l.readBufferBytes > 0 {
		_ = unixConn.SetReadBuffer(l.readBufferBytes)
	}
	if l.writeBufferBytes > 0 {
		_ = unixConn.SetWriteBuffer(l.writeBufferBytes)
	}

	return conn, nil
}
