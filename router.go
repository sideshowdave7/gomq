package gomq

import (
	"github.com/zeromq/gomq/zmtp"
	"net"
)

// RouterSocket is a ZMQ_ROUTER socket type.
// See: http://rfc.zeromq.org/spec:41
type RouterSocket struct {
	*Socket
}

// NewRouter accepts a zmtp.SecurityMechanism and returns
// a RouterSocket as a gomq.Router interface.
func NewRouter(mechanism zmtp.SecurityMechanism, identity string) Router {
	return &RouterSocket{
		Socket: NewSocket(false, zmtp.RouterSocketType, mechanism, zmtp.SocketIdentity(identity)),
	}
}

// Connect accepts a zeromq endpoint and connects the
// router socket to it. Currently the only transport
// supported is TCP. The endpoint string should be
// in the format "tcp://<address>:<port>".
func (r *RouterSocket) Bind(endpoint string) (net.Addr, error) {
	return BindRouter(r, endpoint)
}
