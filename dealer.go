package gomq

import "github.com/zeromq/gomq/zmtp"

// DealerSocket is a ZMQ_DEALER socket type.
// See: http://rfc.zeromq.org/spec:41
type DealerSocket struct {
	*Socket
}

// NewDealer accepts a zmtp.SecurityMechanism and returns
// a DealerSocket as a gomq.Dealer interface.
func NewDealer(mechanism zmtp.SecurityMechanism) Dealer {
	return &DealerSocket{
		Socket: NewSocket(false, zmtp.DealerSocketType, mechanism),
	}
}

// Connect accepts a zeromq endpoint and connects the
// dealer socket to it. Currently the only transport
// supported is TCP. The endpoint string should be
// in the format "tcp://<address>:<port>".
func (c *DealerSocket) Connect(endpoint string) error {
	return ConnectDealer(c, endpoint)
}
