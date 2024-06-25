package tracing

import (
	"net"
)

type Connection struct {
	net.Conn
	OnEventCallback func(clientClosed bool, serverClosed bool, err error)
}

func (t *Connection) CallEvent(err error) {
	if t.OnEventCallback == nil {
		return
	}
	if err == nil {
		return
	}
	/*
		switch {
		case
			errors.Is(err, net.ErrClosed),
			errors.Is(err, io.EOF),
			errors.Is(err, syscall.EPIPE):
			t.OnEventCallback(false, true, err)
		default:
			t.OnEventCallback(false, false, err)
		}
	*/
}

func (t *Connection) Read(b []byte) (n int, err error) {
	n, err = t.Conn.Read(b)
	if err != nil {
		t.CallEvent(err)
	}
	return
}

func (t *Connection) Write(b []byte) (n int, err error) {
	n, err = t.Conn.Write(b)
	if err != nil {
		t.CallEvent(err)
	}
	return
}

func (t *Connection) Close() error {
	err := t.Conn.Close()
	if t.OnEventCallback != nil {
		t.OnEventCallback(true, false, err)
	}
	return err
}
