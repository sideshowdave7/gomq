package gomq

import (
	"net"
	"strings"
	"time"

	"github.com/zeromq/gomq/zmtp"
)

var (
	defaultRetry = 250 * time.Millisecond
)

// Connection is a gomq connection. It holds
// both the net.Conn transport as well as the
// zmtp connection information.
type Connection struct {
	net  net.Conn
	zmtp *zmtp.Connection
}

// ZeroMQSocket is the base gomq interface.
type ZeroMQSocket interface {
	Recv() ([]byte, error)
	Send([]byte) error
	SendMultipart([][]byte) error
	SendMultipartString([]string) error
	RetryInterval() time.Duration
	SocketType() zmtp.SocketType
	SocketIdentity() zmtp.SocketIdentity
	SecurityMechanism() zmtp.SecurityMechanism
	AddConnection(*Connection)
	AddConnectionWithUUID(*Connection, string)
	RemoveConnection(string)
	RecvChannel() chan *zmtp.Message
	RecvMultipart() ([][]byte, error)
	Close()
}

// Client is a gomq interface used for client sockets.
// It implements the Socket interface along with a
// Connect method for connecting to endpoints.
type Client interface {
	ZeroMQSocket
	Connect(endpoint string) error
}

// ConnectClient accepts a Client interface and an endpoint
// in the format <proto>://<address>:<port>. It then attempts
// to connect to the endpoint and perform a ZMTP handshake.
func ConnectClient(c Client, endpoint string) error {
	parts := strings.Split(endpoint, "://")

Connect:
	netConn, err := net.Dial(parts[0], parts[1])
	if err != nil {
		time.Sleep(c.RetryInterval())
		goto Connect
	}

	zmtpConn := zmtp.NewConnection(netConn)
	_, err = zmtpConn.Prepare(c.SecurityMechanism(), c.SocketType(), c.SocketIdentity(), false, nil)
	if err != nil {
		return err
	}

	conn := &Connection{
		net:  netConn,
		zmtp: zmtpConn,
	}

	c.AddConnection(conn)
	zmtpConn.Recv(c.RecvChannel())
	return nil
}

// Server is a gomq interface used for server sockets.
// It implements the Socket interface along with a
// Bind method for binding to endpoints.
type Server interface {
	ZeroMQSocket
	Bind(endpoint string) (net.Addr, error)
}

// BindServer accepts a Server interface and an endpoint
// in the format <proto>://<address>:<port>. It then attempts
// to bind to the endpoint. TODO: change this to starting
// a listener on the endpoint that performs handshakes
// with any client that connects
func BindServer(s Server, endpoint string) (net.Addr, error) {
	var addr net.Addr
	parts := strings.Split(endpoint, "://")

	ln, err := net.Listen(parts[0], parts[1])
	if err != nil {
		return addr, err
	}

	netConn, err := ln.Accept()
	if err != nil {
		return addr, err
	}

	zmtpConn := zmtp.NewConnection(netConn)
	_, err = zmtpConn.Prepare(s.SecurityMechanism(), s.SocketType(), s.SocketIdentity(), true, nil)
	if err != nil {
		return netConn.LocalAddr(), err
	}

	conn := &Connection{
		net:  netConn,
		zmtp: zmtpConn,
	}

	s.AddConnection(conn)
	zmtpConn.Recv(s.RecvChannel())
	return netConn.LocalAddr(), nil
}

type Dealer interface {
	ZeroMQSocket
	Connect(endpoint string) error
}

// ConnectDealer accepts a Dealer interface and an endpoint
// in the format <proto>://<address>:<port>. It then attempts
// to connect to the endpoint and perform a ZMTP handshake.
func ConnectDealer(d Dealer, endpoint string) error {
	parts := strings.Split(endpoint, "://")

Connect:
	netConn, err := net.Dial(parts[0], parts[1])
	if err != nil {
		time.Sleep(d.RetryInterval())
		goto Connect
	}

	zmtpConn := zmtp.NewConnection(netConn)

	_, err = zmtpConn.Prepare(d.SecurityMechanism(), d.SocketType(), d.SocketIdentity(), false, nil)

	if err != nil {
		return err
	}

	conn := &Connection{
		net:  netConn,
		zmtp: zmtpConn,
	}

	d.AddConnection(conn)
	zmtpConn.RecvMultipart(d.RecvChannel())
	return nil
}

type Router interface {
	ZeroMQSocket
	Bind(endpoint string) (net.Addr, error)
}

// BindRouter accepts a Router interface and an endpoint
// in the format <proto>://<address>:<port>. It then attempts
// to bind to the endpoint.
func BindRouter(r Router, endpoint string) (net.Addr, error) {
	var addr net.Addr
	parts := strings.Split(endpoint, "://")

	ln, err := net.Listen(parts[0], parts[1])
	if err != nil {
		return addr, err
	}

	netConn, err := ln.Accept()
	if err != nil {
		return addr, err
	}

	zmtpConn := zmtp.NewConnection(netConn)
	otherEndApplicationMetadata := make(map[string]string)

	otherEndApplicationMetadata, err = zmtpConn.Prepare(r.SecurityMechanism(), r.SocketType(), r.SocketIdentity(), true, nil)
	if err != nil {
		return netConn.LocalAddr(), err
	}

	conn := &Connection{
		net:  netConn,
		zmtp: zmtpConn,
	}

	r.AddConnectionWithUUID(conn, otherEndApplicationMetadata["Identity"])
	zmtpConn.RecvMultipart(r.RecvChannel())
	return netConn.LocalAddr(), nil
}
