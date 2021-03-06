package link

import (
	"sync"
	"time"
)

// Broadcaster.
type Broadcaster struct {
	protocol ProtocolState
	fetcher  func(func(SessionAble))
}

// Broadcast work.
type BroadcastWork struct {
	Session SessionAble
	AsyncWork
}

// Create a broadcaster.
func NewBroadcaster(protocol ProtocolState, fetcher func(func(SessionAble))) *Broadcaster {
	return &Broadcaster{
		protocol: protocol,
		fetcher:  fetcher,
	}
}

// Broadcast to sessions. The message only encoded once
// so the performance is better than send message one by one.
func (b *Broadcaster) Broadcast(message Message, timeout time.Duration) ([]BroadcastWork, error) {
	buffer := NewOutBuffer()

	if err := b.protocol.WriteToBuffer(&buffer, message); err != nil {
		// buffer.free()
		return nil, err
	}
	works := make([]BroadcastWork, 0, 10)
	b.fetcher(func(session SessionAble) {
		// buffer.broadcastUse()
		works = append(works, BroadcastWork{
			session,
			session.AsyncSendBuffer(&buffer, timeout),
		})
	})
	return works, nil
}

// The channel type. Used to maintain a group of session.
// Normally used for broadcast classify purpose.
type Channel struct {
	mutex       sync.RWMutex
	sessions    map[uint64]channelSession
	broadcaster *Broadcaster

	// channel state
	State interface{}
}

type channelSession struct {
	SessionAble
	KickCallback func()
}

// Create a channel instance.
func NewChannel(protocol Protocol, side ProtocolSide) *Channel {
	channel := &Channel{
		sessions: make(map[uint64]channelSession),
	}
	protocolState, _ := protocol.New(channel, side)
	channel.broadcaster = NewBroadcaster(protocolState, channel.Fetch)
	return channel
}

// Broadcast to channel. The message only encoded once
// so the performance is better than send message one by one.
func (channel *Channel) Broadcast(message Message, timeout time.Duration) ([]BroadcastWork, error) {
	return channel.broadcaster.Broadcast(message, timeout)
}

// How mush sessions in this channel.
func (channel *Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	return len(channel.sessions)
}

// Join the channel. The kickCallback will called when the session kick out from the channel.
func (channel *Channel) Join(session SessionAble, kickCallback func()) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.AddCloseCallback(channel, func() {
		channel.Exit(session)
	})
	channel.sessions[session.Id()] = channelSession{session, kickCallback}
}

// Exit the channel.
func (channel *Channel) Exit(session SessionAble) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.RemoveCloseCallback(channel)
	delete(channel.sessions, session.Id())
}

// Kick out a session from the channel.
func (channel *Channel) Kick(sessionId uint64) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	if session, exists := channel.sessions[sessionId]; exists {
		delete(channel.sessions, sessionId)
		if session.KickCallback != nil {
			session.KickCallback()
		}
	}
}

// Fetch the sessions. NOTE: Invoke Kick() or Exit() in fetch callback will dead lock.
func (channel *Channel) Fetch(callback func(SessionAble)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	for _, sesssion := range channel.sessions {
		callback(sesssion.SessionAble)
	}
}
