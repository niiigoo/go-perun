// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
	iotest "perun.network/go-perun/pkg/io/test"
)

func TestMockApp(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))

	address := wallet.NewRandomAddress(rng)
	app := NewMockApp(address)

	t.Run("App", func(t *testing.T) {
		assert.Equal(t, address, app.Def())
	})

	t.Run("StateApp", func(t *testing.T) {
		MockStateAppTest(t, *app)
	})

	t.Run("ActionApp", func(t *testing.T) {
		MockActionAppTest(t, *app)
	})

	t.Run("GenericSerializeable", func(t *testing.T) {
		iotest.GenericSerializableTest(t, NewMockOp(OpValid))
		iotest.GenericSerializableTest(t, NewMockOp(OpErr))
		iotest.GenericSerializableTest(t, NewMockOp(OpTransitionErr))
		iotest.GenericSerializableTest(t, NewMockOp(OpActionErr))
		iotest.GenericSerializableTest(t, NewMockOp(OpPanic))
	})

	// We cant use VerifyClone here since it requires that the same type is returned by
	// Clone() but in this case it returns channel.Data instead of *MockOp
	t.Run("CloneTest", func(t *testing.T) {
		op := NewMockOp(OpValid)
		op2 := op.Clone()
		// Dont use Equal here since it compares the values and not addresses
		assert.False(t, op == op2, "Clone should return a different address")
		assert.Equal(t, op, op2, "Clone should return the same value")
	})
}

func MockStateAppTest(t *testing.T, app MockApp) {
	stateValid := createState(OpValid)
	stateErr := createState(OpErr)
	stateTransErr := createState(OpTransitionErr)
	stateActErr := createState(OpActionErr)
	statePanic := createState(OpPanic)

	t.Run("ValidTransition", func(t *testing.T) {
		// ValidTransition only checks the first state.
		assert.NoError(t, app.ValidTransition(nil, stateValid, nil))
		assert.Error(t, app.ValidTransition(nil, stateErr, nil))
		assert.True(t, channel.IsStateTransitionError(app.ValidTransition(nil, stateTransErr, nil)))
		assert.True(t, channel.IsActionError(app.ValidTransition(nil, stateActErr, nil)))
		assert.Panics(t, func() { assert.NoError(t, app.ValidTransition(nil, statePanic, nil)) })
	})

	t.Run("ValidInit", func(t *testing.T) {
		assert.NoError(t, app.ValidInit(nil, stateValid))
		assert.Error(t, app.ValidInit(nil, stateErr))
		assert.True(t, channel.IsStateTransitionError(app.ValidInit(nil, stateTransErr)))
		assert.True(t, channel.IsActionError(app.ValidInit(nil, stateActErr)))
		assert.Panics(t, func() { assert.NoError(t, app.ValidInit(nil, statePanic)) })
	})
}

func MockActionAppTest(t *testing.T, app MockApp) {
	actValid := NewMockOp(OpValid)
	actErr := NewMockOp(OpErr)
	actTransErr := NewMockOp(OpTransitionErr)
	actActErr := NewMockOp(OpActionErr)
	actPanic := NewMockOp(OpPanic)

	state := createState(OpValid)

	t.Run("InitState", func(t *testing.T) {
		_, _, err := app.InitState(nil, []channel.Action{actValid})
		// Sadly we can not check Allocation.valid() here, since it is private.
		assert.NoError(t, err)

		_, _, err = app.InitState(nil, []channel.Action{actErr})
		assert.Error(t, err)

		_, _, err = app.InitState(nil, []channel.Action{actTransErr})
		assert.True(t, channel.IsStateTransitionError(err))

		_, _, err = app.InitState(nil, []channel.Action{actActErr})
		assert.True(t, channel.IsActionError(err))

		assert.Panics(t, func() { app.InitState(nil, []channel.Action{actPanic}) })
	})

	t.Run("ValidAction", func(t *testing.T) {
		assert.NoError(t, app.ValidAction(nil, nil, 0, actValid))
		assert.Error(t, app.ValidAction(nil, nil, 0, actErr))
		assert.True(t, channel.IsStateTransitionError(app.ValidAction(nil, nil, 0, actTransErr)))
		assert.True(t, channel.IsActionError(app.ValidAction(nil, nil, 0, actActErr)))
		assert.Panics(t, func() { app.ValidAction(nil, nil, 0, actPanic) })
	})

	t.Run("ApplyActions", func(t *testing.T) {
		// ApplyActions increments the Version counter, so we cant pass nil as state.
		retState, err := app.ApplyActions(nil, state, []channel.Action{actValid})
		assert.Equal(t, retState.Version, state.Version+1)
		assert.NoError(t, err)

		_, err = app.ApplyActions(nil, state, []channel.Action{actErr})
		assert.Error(t, err)

		_, err = app.ApplyActions(nil, state, []channel.Action{actTransErr})
		assert.True(t, channel.IsStateTransitionError(err))

		_, err = app.ApplyActions(nil, state, []channel.Action{actActErr})
		assert.True(t, channel.IsActionError(err))

		assert.Panics(t, func() { app.ApplyActions(nil, state, []channel.Action{actPanic}) })
	})
}

func createState(op MockOp) *channel.State {
	return &channel.State{ID: channel.ID{}, Version: 0, Allocation: channel.Allocation{}, Data: NewMockOp(op), IsFinal: false}
}