// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"context"
	"net"

	"github.com/pkg/errors"

	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sync/atomic"
)

var _ peer.Dialer = (*Dialer)(nil)

// Dialer is a test dialer that can dial connections to Listeners via a ConnHub.
type Dialer struct {
	hub      *ConnHub
	identity peer.Identity

	closed atomic.Bool
}

func (d *Dialer) Dial(ctx context.Context, address peer.Address) (peer.Conn, error) {
	if d.closed.IsSet() {
		return nil, errors.New("dialer closed")
	}

	select {
	case <-ctx.Done():
		return nil, errors.New("manually aborted")
	default:
	}

	if l, ok := d.hub.find(address); ok {
		local, remote := net.Pipe()
		l.Put(peer.NewIoConn(remote))
		conn := peer.NewIoConn(local)
		if addr, err := peer.ExchangeAddrs(d.identity, conn); err != nil {
			return nil, err
		} else if !addr.Equals(address) {
			return nil, errors.New("invalid peer address")
		}
		return conn, nil
	}

	return nil, errors.Errorf("peer with address %v not found", address)
}

func (d *Dialer) Close() error {
	if !d.closed.TrySet() {
		return errors.New("dialer was already closed")
	}
	return nil
}