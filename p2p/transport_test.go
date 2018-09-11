package p2p

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestTransportMultiplexConnFilter(t *testing.T) {
	mt := NewMultiplexTransport(
		NodeInfo{},
		NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)

	MultiplexTransportConnFilters(
		func(_ map[string]net.Conn, _ net.Conn) error { return nil },
		func(_ map[string]net.Conn, _ net.Conn) error { return nil },
		func(_ map[string]net.Conn, _ net.Conn) error {
			return fmt.Errorf("rejected")
		},
	)(mt)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err = mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsFiltered() {
			t.Errorf("expected peer to be filtered")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestTransportMultiplexConnFilterTimeout(t *testing.T) {
	mt := NewMultiplexTransport(
		NodeInfo{},
		NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)

	MultiplexTransportFilterTimeout(5 * time.Millisecond)(mt)
	MultiplexTransportConnFilters(
		func(_ map[string]net.Conn, _ net.Conn) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	)(mt)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err = mt.Accept(peerConfig{})
	if errors.Cause(err) != ErrTransportFilterTimeout {
		t.Errorf("expected ErrTransportFilterTimeout")
	}
}

func TestTransportMultiplexPeerFilter(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	MultiplexTransportPeerFilters(
		func(_ map[ID]Peer, _ Peer) error { return nil },
		func(_ map[ID]Peer, _ Peer) error {
			fmt.Println("THE FAK")
			return fmt.Errorf("ID is not welcome here")
		},
		func(_ map[ID]Peer, _ Peer) error { return nil },
	)(mt)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		var (
			pv     = ed25519.GenPrivKey()
			dialer = NewMultiplexTransport(
				NodeInfo{
					ID:         PubKeyToID(pv.PubKey()),
					ListenAddr: "127.0.0.1:0",
					Moniker:    "dialer",
					Version:    "1.0.0",
				},
				NodeKey{
					PrivKey: pv,
				},
			)
		)

		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err = mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsFiltered() {
			t.Errorf("expected peer to be filtered")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestTransportMultiplexPeerFilterTimeout(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	MultiplexTransportFilterTimeout(5 * time.Millisecond)(mt)
	MultiplexTransportPeerFilters(
		func(_ map[ID]Peer, _ Peer) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	)(mt)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		var (
			pv     = ed25519.GenPrivKey()
			dialer = NewMultiplexTransport(
				NodeInfo{
					ID:         PubKeyToID(pv.PubKey()),
					ListenAddr: "127.0.0.1:0",
					Moniker:    "dialer",
					Version:    "1.0.0",
				},
				NodeKey{
					PrivKey: pv,
				},
			)
		)

		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err = mt.Accept(peerConfig{})
	if errors.Cause(err) != ErrTransportFilterTimeout {
		t.Errorf("expected ErrTransportFilterTimeout")
	}
}

func TestTransportMultiplexPeerFilterDuplicate(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		var (
			pv     = ed25519.GenPrivKey()
			dialer = NewMultiplexTransport(
				NodeInfo{
					ID:         PubKeyToID(pv.PubKey()),
					ListenAddr: "127.0.0.1:0",
					Moniker:    "dialer",
					Version:    "1.0.0",
				},
				NodeKey{
					PrivKey: pv,
				},
			)
		)

		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		duplicate := NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "dialer",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)

		_, err = duplicate.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	if _, err := mt.Accept(peerConfig{}); err != nil {
		t.Fatal(err)
	}

	_, err = mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsDuplicate() {
			t.Errorf("expected peer to be duplicate")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestTransportMultiplexAcceptMultiple(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	var (
		seed = rand.New(rand.NewSource(time.Now().UnixNano()))
		errc = make(chan error, seed.Intn(64)+64)
	)

	// Setup dialers.
	for i := 0; i < cap(errc); i++ {
		go func() {
			var (
				pv     = ed25519.GenPrivKey()
				dialer = NewMultiplexTransport(
					NodeInfo{
						ID:         PubKeyToID(pv.PubKey()),
						ListenAddr: "127.0.0.1:0",
						Moniker:    "dialer",
						Version:    "1.0.0",
					},
					NodeKey{
						PrivKey: pv,
					},
				)
			)

			addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
			if err != nil {
				errc <- err
				return
			}

			_, err = dialer.Dial(*addr, peerConfig{})
			if err != nil {
				errc <- err
				return
			}

			// Signal that the connection was established.
			errc <- nil
		}()
	}

	// Catch connection errors.
	for i := 0; i < cap(errc); i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
	}

	ps := []Peer{}

	// Accept all peers.
	for i := 0; i < cap(errc); i++ {
		p, err := mt.Accept(peerConfig{})
		if err != nil {
			t.Fatal(err)
		}

		if err := p.Start(); err != nil {
			t.Fatal(err)
		}

		ps = append(ps, p)
	}

	if have, want := len(ps), cap(errc); have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	if have, want := len(mt.peers), cap(errc); have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	// Stop all peers.
	for _, p := range ps {
		if err := p.Stop(); err != nil {
			t.Fatal(err)
		}
	}

	// Test that we successfully removed peers after its lifecycle is complete.
	if have, want := len(mt.peers), 0; have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	if err := mt.Close(); err != nil {
		t.Errorf("close errored: %v", err)
	}
}

func TestTransportMultiplexAcceptNonBlocking(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	var (
		fastNodePV   = ed25519.GenPrivKey()
		fastNodeInfo = NodeInfo{
			ID:         PubKeyToID(fastNodePV.PubKey()),
			ListenAddr: "127.0.0.1:0",
			Moniker:    "fastNode",
			Version:    "1.0.0",
		}
		errc  = make(chan error)
		fastc = make(chan struct{})
		slowc = make(chan struct{})
	)

	// Simulate slow Peer.
	go func() {
		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		c, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(slowc)

		select {
		case <-fastc:
			// Fast peer connected.
		case <-time.After(50 * time.Millisecond):
			// We error if the fast peer didn't succeed.
			errc <- fmt.Errorf("Fast peer timed out")
		}

		sc, err := secretConn(c, 20*time.Millisecond, ed25519.GenPrivKey())
		if err != nil {
			errc <- err
			return
		}

		_, err = handshake(sc, 20*time.Millisecond, NodeInfo{
			ID:         PubKeyToID(ed25519.GenPrivKey().PubKey()),
			ListenAddr: "127.0.0.1:0",
			Moniker:    "slow_peer",
		})
		if err != nil {
			errc <- err
			return
		}
	}()

	// Simulate fast Peer.
	go func() {
		<-slowc

		var (
			dialer = NewMultiplexTransport(
				fastNodeInfo,
				NodeKey{
					PrivKey: fastNodePV,
				},
			)
		)

		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
		close(fastc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	p, err := mt.Accept(peerConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := p.NodeInfo(), fastNodeInfo; !reflect.DeepEqual(have, want) {
		t.Errorf("have %v, want %v", have, want)
	}
}

func TestTransportMultiplexValidateNodeInfo(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		var (
			pv     = ed25519.GenPrivKey()
			dialer = NewMultiplexTransport(
				NodeInfo{
					ID:         PubKeyToID(pv.PubKey()),
					ListenAddr: "127.0.0.1:0",
					Moniker:    "", // Should not be empty.
					Version:    "1.0.0",
				},
				NodeKey{
					PrivKey: pv,
				},
			)
		)

		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err = mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsNodeInfoInvalid() {
			t.Errorf("expected NodeInfo to be invalid")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestTransportMultiplexRejectMissmatchID(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		dialer := NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(ed25519.GenPrivKey().PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "dialer",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: ed25519.GenPrivKey(),
			},
		)

		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err = mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsAuthFailure() {
			t.Errorf("expected auth failure")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestTransportMultiplexRejectIncompatible(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		var (
			pv     = ed25519.GenPrivKey()
			dialer = NewMultiplexTransport(
				NodeInfo{
					ID:         PubKeyToID(pv.PubKey()),
					ListenAddr: "127.0.0.1:0",
					Moniker:    "dialer",
					Version:    "2.0.0",
				},
				NodeKey{
					PrivKey: pv,
				},
			)
		)

		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	_, err = mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsIncompatible() {
			t.Errorf("expected to reject incompatible")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestTransportMultiplexRejectSelf(t *testing.T) {
	var (
		pv = ed25519.GenPrivKey()
		mt = NewMultiplexTransport(
			NodeInfo{
				ID:         PubKeyToID(pv.PubKey()),
				ListenAddr: "127.0.0.1:0",
				Moniker:    "transport",
				Version:    "1.0.0",
			},
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressStringWithOptionalID("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		addr, err := NewNetAddressStringWithOptionalID(mt.listener.Addr().String())
		if err != nil {
			errc <- err
			return
		}

		_, err = mt.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		if err, ok := err.(ErrRejected); ok {
			if !err.IsSelf() {
				t.Errorf("expected to reject self")
			}
		} else {
			t.Errorf("expected ErrRejected")
		}
	} else {
		t.Errorf("expected connection failure")
	}

	_, err = mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsSelf() {
			t.Errorf("expected to reject self")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestTransportConnDuplicateIPFilter(t *testing.T) {
	filter := ConnDuplicateIPFilter(&testTransportResolver{})

	if err := filter(nil, &testTransportConn{}); err != nil {
		t.Fatal(err)
	}

	var (
		c  = &testTransportConn{}
		cs = map[string]net.Conn{c.RemoteAddr().String(): c}
	)

	if err := filter(cs, c); err == nil {
		t.Errorf("expected Peer to be rejected as duplicatge")
	}
}

func TestTransportHandshake(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var (
		peerPV       = ed25519.GenPrivKey()
		peerNodeInfo = NodeInfo{
			ID: PubKeyToID(peerPV.PubKey()),
		}
	)

	go func() {
		c, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			t.Error(err)
			return
		}

		go func(c net.Conn) {
			_, err := cdc.MarshalBinaryWriter(c, peerNodeInfo)
			if err != nil {
				t.Error(err)
			}
		}(c)
		go func(c net.Conn) {
			ni := NodeInfo{}

			_, err := cdc.UnmarshalBinaryReader(
				c,
				&ni,
				int64(MaxNodeInfoSize()),
			)
			if err != nil {
				t.Error(err)
			}
		}(c)
	}()

	c, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}

	ni, err := handshake(c, 20*time.Millisecond, NodeInfo{})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := ni, peerNodeInfo; !reflect.DeepEqual(have, want) {
		t.Errorf("have %v, want %v", have, want)
	}
}

type testTransportAddr struct{}

func (a *testTransportAddr) Network() string { return "tcp" }
func (a *testTransportAddr) String() string  { return "test.local:1234" }

type testTransportConn struct{}

func (c *testTransportConn) Close() error {
	return fmt.Errorf("Close() not implemented")
}

func (c *testTransportConn) LocalAddr() net.Addr {
	return &testTransportAddr{}
}

func (c *testTransportConn) RemoteAddr() net.Addr {
	return &testTransportAddr{}
}

func (c *testTransportConn) Read(_ []byte) (int, error) {
	return -1, fmt.Errorf("Read() not implemented")
}

func (c *testTransportConn) SetDeadline(_ time.Time) error {
	return fmt.Errorf("SetDeadline() not implemented")
}

func (c *testTransportConn) SetReadDeadline(_ time.Time) error {
	return fmt.Errorf("SetReadDeadline() not implemented")
}

func (c *testTransportConn) SetWriteDeadline(_ time.Time) error {
	return fmt.Errorf("SetWriteDeadline() not implemented")
}

func (c *testTransportConn) Write(_ []byte) (int, error) {
	return -1, fmt.Errorf("Write() not implemented")
}

type testTransportResolver struct{}

func (r *testTransportResolver) LookupIPAddr(
	ctx context.Context,
	host string,
) ([]net.IPAddr, error) {

	return []net.IPAddr{
		{
			IP: net.IP{10, 0, 10, 1},
		},
		{
			IP: net.IP{10, 0, 10, 2},
		},
		{
			IP: net.IP{10, 0, 10, 3},
		},
	}, nil
}
