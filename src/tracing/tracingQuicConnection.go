package tracing

import (
	"context"
	"fmt"
	"github.com/lucas-clemente/quic-go"
)

type TracingQuicConnection struct {
	quic.EarlyConnection
	openedOutgoingStreams  int
	currentOutgoingStreams int
	closedOutgoingStreams  int
	openedIncomingStreams  int
	currentIncomingStreams int
	closedIncomingStreams  int
}

type TracingQuicSendStream struct {
	quic.SendStream
}

type TracingQuicReceiveStream struct {
	quic.ReceiveStream
}

func (t TracingQuicConnection) OpenStream() (quic.Stream, error) {
	panic("not implemented")
}

func (t TracingQuicConnection) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	fmt.Println("OpenStreamSync")
	return t.EarlyConnection.OpenStreamSync(ctx)
}

func (t TracingQuicConnection) OpenUniStream() (quic.SendStream, error) {
	fmt.Println("OpenUniStream")
	t.openedOutgoingStreams++
	t.currentOutgoingStreams++
	sendStream, err := t.EarlyConnection.OpenUniStream()
	return TracingQuicSendStream{sendStream}, err
}

func (t TracingQuicConnection) OpenUniStreamSync(context.Context) (quic.SendStream, error) {
	panic("not implemented")
}
func (t TracingQuicConnection) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	fmt.Println("AcceptUniStream")
	t.openedIncomingStreams++
	t.currentIncomingStreams++
	readStream, err := t.EarlyConnection.AcceptUniStream(ctx)
	return TracingQuicReceiveStream{readStream}, err
}

func (t TracingQuicSendStream) Write(p []byte) (n int, err error) {
	fmt.Println("Write")
	fmt.Println(p)

	return t.SendStream.Write(p)
}

func (t TracingQuicReceiveStream) Read(p []byte) (n int, err error) {
	fmt.Println("Read")
	n, err = t.ReceiveStream.Read(p)
	fmt.Println(p)
	return
}
