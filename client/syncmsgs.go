// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"io"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wire"
)

func init() {
	wire.RegisterDecoder(wire.ChannelSync,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelSync
			return &m, m.Decode(r)
		})
}

type msgChannelSync struct {
	Phase     channel.Phase       // Phase is the phase of the sender.
	CurrentTX channel.Transaction // CurrentTX is the sender's current transaction.
}

var _ ChannelMsg = (*msgChannelSync)(nil)

func newChannelSyncMsg(s channel.Source) *msgChannelSync {
	return &msgChannelSync{
		Phase:     s.Phase(),
		CurrentTX: s.CurrentTX(),
	}
}

// Encode implements perunio.Encode.
func (m *msgChannelSync) Encode(w io.Writer) error {
	return perunio.Encode(w,
		m.Phase,
		m.CurrentTX)
}

// Decode implements perunio.Decode.
func (m *msgChannelSync) Decode(r io.Reader) error {
	return perunio.Decode(r,
		&m.Phase,
		&m.CurrentTX)
}

// ID returns the channel's ID.
func (m *msgChannelSync) ID() channel.ID {
	return m.CurrentTX.ID
}

// Type implements wire.Type.
func (m *msgChannelSync) Type() wire.Type {
	return wire.ChannelSync
}
