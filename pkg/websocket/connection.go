package websocket

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 2700
)

type Connection struct {
	// The websocket connection.
	conn *websocket.Conn

	messageHandler messageHandler

	// Buffered channel of outbound messages.
	send chan []byte

	closeConn chan struct{}

	onClose    func()
	CloseEvent chan struct{}
}

// in - request data,
// f - function to send out as a response
type messageHandler func(in []byte, f func(out []byte))

// RW is abbr for Request & responseWriter
type RW struct {
	W http.ResponseWriter
	R *http.Request
}

func NewConnection(rw *RW, handler messageHandler, onClose func()) (*Connection, error) {
	conn, err := upgrader.Upgrade(rw.W, rw.R, nil)
	if err != nil {
		return nil, err
	}
	client := &Connection{
		conn:           conn,
		send:           make(chan []byte, 256),
		messageHandler: handler,
		closeConn:      make(chan struct{}),
		CloseEvent:     make(chan struct{}),
		onClose:        onClose,
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	return client, nil
}

func (c *Connection) readPump() {
	defer func() {
		close(c.closeConn)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.refreshReadDeadline()

	c.conn.SetPongHandler(c.pong)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket: %v\n", err)
			}
			break
		}
		message = bytes.ReplaceAll(message, newline, space)
		c.messageHandler(message, c.Send)
	}
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		if err != nil {
			log.Println(err)
		}

		close(c.CloseEvent)
		if c.onClose != nil {
			c.onClose()
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Printf("writeDeadline error: %v\n", err)
			}
			if !ok {
				// The hub closed the channel.
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("writeMessage error: %v\n", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("next writer error: %v\n", err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				log.Printf("write error: %v\n", err)
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err := w.Write(newline)
				if err != nil {
					log.Printf("write error: %v\n", err)
					return
				}

				_, err = w.Write(<-c.send)
				if err != nil {
					log.Printf("write error: %v\n", err)
					return
				}
			}

			if err := w.Close(); err != nil {
				log.Printf("write Close error: %v\n", err)
				return
			}
		case <-ticker.C:
			c.ping()
		case <-c.closeConn:
			return
		}
	}
}

func (c *Connection) ping() {
	err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		log.Printf("writeDeadline error: %v\n", err)
	}

	err = c.conn.WriteMessage(websocket.PingMessage, nil)
	if err != nil {
		log.Printf("ping error: %v\n", err)
		return
	}
}

func (c *Connection) pong(_ string) error {
	c.refreshReadDeadline()
	return nil
}

func (c *Connection) refreshReadDeadline() {
	e := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if e != nil {
		log.Println(e)
	}
}

func (c *Connection) Send(message []byte) {
	c.send <- message
}

func (c *Connection) Close() {
	close(c.send)
}
