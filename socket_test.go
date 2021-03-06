package gomq

import (
	"bytes"
	"net"
	"testing"

	"github.com/zeromq/gomq/internal/test"
	"github.com/zeromq/gomq/zmtp"
)

func TestNewClient(t *testing.T) {
	var addr net.Addr
	var err error

	go func() {
		client := NewClient(zmtp.NewSecurityNull())
		err = client.Connect("tcp://127.0.0.1:9999")
		if err != nil {
			t.Error(err)
		}

		err := client.Send([]byte("HELLO"))
		if err != nil {
			t.Error(err)
		}

		msg, _ := client.Recv()
		if want, got := 0, bytes.Compare([]byte("WORLD"), msg); want != got {
			t.Errorf("want %v, got %v", want, got)
		}

		t.Logf("client received: %q", string(msg))

		err = client.Send([]byte("GOODBYE"))
		if err != nil {
			t.Error(err)
		}

		client.Close()
	}()

	server := NewServer(zmtp.NewSecurityNull())

	addr, err = server.Bind("tcp://127.0.0.1:9999")
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "127.0.0.1:9999", addr.String(); want != got {
		t.Fatalf("want %q, got %q", want, got)
	}

	if err != nil {
		t.Fatal(err)
	}

	msg, _ := server.Recv()

	if want, got := 0, bytes.Compare([]byte("HELLO"), msg); want != got {
		t.Fatalf("want %q, got %q", []byte("HELLO"), msg)
	}

	t.Logf("server received: %q", string(msg))

	server.Send([]byte("WORLD"))

	msg, err = server.Recv()
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 0, bytes.Compare([]byte("GOODBYE"), msg); want != got {
		t.Fatalf("want %q, got %q", []byte("GOODBYE"), msg)
	}

	t.Logf("server received: %q", string(msg))

	server.Close()
}

func TestExternalServer(t *testing.T) {
	t.Logf("Testing Server")
	go test.StartExternalServer()

	client := NewClient(zmtp.NewSecurityNull())
	err := client.Connect("tcp://127.0.0.1:31337")
	if err != nil {
		t.Fatal(err)
	}

	err = client.Send([]byte("HELLO"))
	if err != nil {
		t.Fatal(err)
	}

	msg, err := client.Recv()
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 0, bytes.Compare([]byte("WORLD"), msg); want != got {
		t.Errorf("want %q, got %q", []byte("WORLD"), msg)
	}

	t.Logf("client received: %q", string(msg))

	client.Close()
}

func TestExternalRouter(t *testing.T) {

	go test.StartExternalRouter()

	dealer := NewDealer(zmtp.NewSecurityNull(), "test_dealer")
	err := dealer.Connect("tcp://127.0.0.1:31340")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Sending HELLO")
	err = dealer.SendMultipartString([]string{"", "HELLO"})
	if err != nil {
		t.Fatal(err)
	}

	msg, _ := dealer.Recv()

	if want, got := 0, bytes.Compare([]byte("WORLD"), msg); want != got {
		t.Errorf("want %v, got %v", want, got)
	}

	t.Logf("dealer received: %q", string(msg))

	t.Log("Sending multipart")
	err = dealer.SendMultipartString([]string{"", "HELLO", "WORLD"})
	if err != nil {
		t.Fatal(err)
	}

	rmsg, _ := dealer.RecvMultipart()

	for i := range rmsg {
		t.Logf("dealer received: %q", string(rmsg[i]))
	}

	if want, got := 0, bytes.Compare([]byte("WORLD"), rmsg[0]); want != got {
		t.Errorf("want %q, got %q", []byte("WORLD"), rmsg[0])
	}

	if want, got := 0, bytes.Compare([]byte("HELLO"), rmsg[1]); want != got {
		t.Errorf("want %q, got %q", []byte("HELLO"), rmsg[1])
	}

	dealer.Close()
}

func TestDealerRouter(t *testing.T) {
	var addr net.Addr
	var err error

	go func() {
		dealer := NewDealer(zmtp.NewSecurityNull(), "test_dealer")
		defer dealer.Close()
		err = dealer.Connect("tcp://127.0.0.1:11123")
		if err != nil {
			t.Fatal(err)
		}

		err = dealer.SendMultipartString([]string{"", "GOODBYE"})
		if err != nil {
			t.Fatal(err)
		}

		_, err := dealer.RecvMultipart()
		t.Logf("test")
		if err != nil {
			t.Log(err)
		}

		// if want, got := 0, bytes.Compare([]byte("HELLO"), msg2[1]); want != got {
		// t.Fatalf("want %v, got %v", want, got)
		// }

		// t.Logf("dealer received: %q", string(msg2[1]))
		dealer.Close()
	}()

	router := NewRouter(zmtp.NewSecurityNull(), "router")
	defer router.Close()

	addr, err = router.Bind("tcp://127.0.0.1:11123")
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "127.0.0.1:11123", addr.String(); want != got {
		t.Fatalf("want %q, got %q", want, got)
	}

	if err != nil {
		t.Fatal(err)
	}

	msg, err := router.RecvMultipart()
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 0, bytes.Compare([]byte("GOODBYE"), msg[2]); want != got {
		t.Fatalf("want %v, got %v (%v)", want, got, msg[2])
	}

	t.Logf("router received: %q", string(msg[2]))

	router.SendMultipartString([]string{string(msg[0]), "", "WORLD"})

	router.Close()
}

func TestPushPull(t *testing.T) {
	var addr net.Addr
	var err error

	go func() {
		pull := NewPull(zmtp.NewSecurityNull())
		defer pull.Close()
		err = pull.Connect("tcp://127.0.0.1:12345")
		if err != nil {
			t.Fatal(err)
		}

		msg, err := pull.Recv()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := 0, bytes.Compare([]byte("HELLO"), msg); want != got {
			t.Fatalf("want %v, got %v", want, got)
		}

		t.Logf("pull received: %q", string(msg))

		err = pull.Send([]byte("GOODBYE"))
		if err != nil {
			t.Fatal(err)
		}

		pull.Close()
	}()

	push := NewPush(zmtp.NewSecurityNull())
	defer push.Close()

	addr, err = push.Bind("tcp://127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "127.0.0.1:12345", addr.String(); want != got {
		t.Fatalf("want %q, got %q", want, got)
	}

	if err != nil {
		t.Fatal(err)
	}

	push.Send([]byte("HELLO"))

	msg, err := push.Recv()
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 0, bytes.Compare([]byte("GOODBYE"), msg); want != got {
		t.Fatalf("want %v, got %v (%v)", want, got, msg)
	}

	t.Logf("push received: %q", string(msg))

	push.Close()
}
