// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet_test

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

var dataToSign = []byte("SomeLongDataThatShouldBeSignedPlease")

const (
	password    = "secret"
	sampleAddr  = "1234560000000000000000000000000000000000"
	invalidAddr = "123456"
)

type Setup struct {
	test.Setup
	Wallet *ethwallet.Wallet
}

func TestGenericSignatureTests(t *testing.T) {
	setup := newSetup(t)
	test.GenericSignatureTest(t, setup)
	test.GenericSignatureSizeTest(t, setup)
}

func TestGenericAddressTests(t *testing.T) {
	test.GenericAddressTest(t, newSetup(t))
}

func TestWallet_Contains(t *testing.T) {
	rng := rand.New(rand.NewSource(0x007))
	w := ethwallettest.NewTmpWallet()

	assert.False(t, w.Contains(test.NewRandomAddress(rng)), "Expected wallet not to contain an empty account")

	acc := w.NewAccount()
	assert.True(t, w.Contains(acc.Address()), "Expected wallet to contain account")
}

func TestSignatures(t *testing.T) {
	acc := ethwallettest.NewTmpWallet().NewAccount()
	sign, err := acc.SignData(dataToSign)
	assert.NoError(t, err, "Sign with new account should succeed")
	assert.Equal(t, len(sign), ethwallet.SigLen, "Ethereum signature has wrong length")
	valid, err := new(ethwallet.Backend).VerifySignature(dataToSign, sign, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.NoError(t, err, "Verification should succeed")
}

func TestBackend(t *testing.T) {
	backend := new(ethwallet.Backend)

	s := newSetup(t)

	buff := bytes.NewReader(s.AddressBytes)
	addr, err := backend.DecodeAddress(buff)

	assert.NoError(t, err, "NewAddress from Bytes should work")
	assert.Equal(t, s.AddressBytes, addr.Bytes())

	buff = bytes.NewReader([]byte(invalidAddr))
	_, err = backend.DecodeAddress(buff)
	assert.Error(t, err, "Conversion from wrong address should fail")
}

func newSetup(t require.TestingT) *test.Setup {
	acc := ethwallettest.NewTmpWallet().NewAccount()
	sampleBytes, err := hex.DecodeString(sampleAddr)
	if err != nil {
		panic("invalid sample address")
	}

	return &test.Setup{
		UnlockedAccount: func() (wallet.Account, error) { return acc, nil },
		Backend:         new(ethwallet.Backend),
		AddressBytes:    sampleBytes,
		DataToSign:      dataToSign,
	}
}

func TestCurve_SigningAndVerifying(t *testing.T) {
	msg, err := hex.DecodeString("f27b90711d11d10a155fc8ba0eed1ffbf449cf3730d88c0cb77b98f61750ab34000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000160000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000010000000000000000000000002c2b9c9a4a25e24b174f26114e8926a9f2128fe40000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000de0b6b3a7640000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err, "decode msg should not error")
	sig, err := hex.DecodeString("538da6430f7915832de165f89c69239020461b80861559a00d4f5a2a7705765219eb3969eb7095f8addb6bf9c9f96f6adf44cfd4a8136516f88b337a428bf1bb1b")
	require.NoError(t, err, "decode sig should not error")
	addr := ethwallet.Address(common.HexToAddress("f17f52151EbEF6C7334FAD080c5704D77216b732"))
	b, err := ethwallet.VerifySignature(msg, sig, &addr)
	assert.NoError(t, err, "VerifySignature should not error")
	assert.True(t, b, "VerifySignature")
}
