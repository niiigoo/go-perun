// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
)

var timeout = 100 * time.Millisecond

// Two nodes (1 dialer, 1 listener node) .Get() each other.
func TestEndpointRegistry_Get_Pair(t *testing.T) {
	t.Parallel()
	assert, require := assert.New(t), require.New(t)
	rng := rand.New(rand.NewSource(3))
	var hub wiretest.ConnHub
	dialerId := wallettest.NewRandomAccount(rng)
	listenerId := wallettest.NewRandomAccount(rng)
	dialerReg := wire.NewEndpointRegistry(dialerId, func(*wire.Endpoint) {}, hub.NewNetDialer())
	listenerReg := wire.NewEndpointRegistry(listenerId, func(*wire.Endpoint) {}, nil)
	listener := hub.NewNetListener(listenerId.Address())

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()
	p, err := dialerReg.Get(ctx, listenerId.Address())
	assert.NoError(err)
	require.NotNil(p)
	assert.True(p.PerunAddress.Equals(listenerId.Address()))

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err = listenerReg.Get(ctx, dialerId.Address())
	assert.NoError(err)
	require.NotNil(p)
	assert.True(p.PerunAddress.Equals(dialerId.Address()))

	assert.NoError(listenerReg.Close())
	assert.NoError(dialerReg.Close())
	assert.True(sync.IsAlreadyClosedError(listener.Close()))
	test.AssertTerminates(t, timeout, func() {
		<-done
	})
}

// Tests that calling .Get() concurrently on the same peer works properly.
func TestEndpointRegistry_Get_Multiple(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	rng := rand.New(rand.NewSource(3))
	var hub wiretest.ConnHub
	dialerId := wallettest.NewRandomAccount(rng)
	listenerId := wallettest.NewRandomAccount(rng)
	dialer := hub.NewNetDialer()
	logPeer := func(p *wire.Endpoint) {
		p.OnCreateAlways(func() { t.Logf("subscribing %x\n", p.PerunAddress.Bytes()[:4]) })
	}
	dialerReg := wire.NewEndpointRegistry(dialerId, logPeer, dialer)
	listenerReg := wire.NewEndpointRegistry(listenerId, logPeer, nil)
	listener := hub.NewNetListener(listenerId.Address())

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()

	const N = 4
	peers := make(chan *wire.Endpoint, N)
	for i := 0; i < N; i++ {
		go func() {
			p, err := dialerReg.Get(ctx, listenerId.Address())
			assert.NoError(err)
			if p != nil {
				assert.True(p.PerunAddress.Equals(listenerId.Address()))
			}
			peers <- p
		}()
	}

	ct := test.NewConcurrent(t)
	test.AssertTerminates(t, timeout, func() {
		ct.Stage("terminates", func(t require.TestingT) {
			require := require.New(t)
			p := <-peers
			require.NotNil(p)
			for i := 0; i < N-1; i++ {
				p0 := <-peers
				require.NotNil(p0)
				assert.Same(p, p0)
			}
		})
	})

	ct.Wait("terminates")

	assert.Equal(1, dialer.NumDialed())

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err := listenerReg.Get(ctx, dialerId.Address())
	assert.NoError(err)
	assert.NotNil(p)
	assert.True(p.PerunAddress.Equals(dialerId.Address()))
	assert.Equal(1, listener.NumAccepted())

	assert.NoError(listenerReg.Close())
	assert.NoError(dialerReg.Close())
	assert.True(sync.IsAlreadyClosedError(listener.Close()))
	test.AssertTerminates(t, timeout, func() { <-done })
}
