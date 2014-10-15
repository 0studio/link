package link

import (
	"net"
)

type MockSession struct {
	id             uint64
	sendPacketChan chan []byte
}

func NewMockSession() *MockSession {
	return &MockSession{}
}

func (Session *MockSession) Start() {
}

func (Session *MockSession) Id() uint64 {
	return Session.id
}

func (Session *MockSession) Conn() net.Conn {
	return nil
}

func (Session *MockSession) IsClosed() bool {
	return false
}

func (Session *MockSession) SyncSendPacket(packet []byte) error {
	return nil
}

func (Session *MockSession) Send(message Message) error {
	return nil
}

func (Session *MockSession) SendPacket(packet []byte) error {
	Session.sendPacketChan <- packet
	return nil
}

func (session *MockSession) Close(reason interface{}) {
}

func (session *MockSession) Read() ([]byte, error) {
	bytes := <-session.sendPacketChan
	return bytes, nil
}
