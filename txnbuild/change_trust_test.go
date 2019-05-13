package txnbuild

import (
	"testing"

	"github.com/stellar/go/network"
	"github.com/stretchr/testify/assert"
)

func TestChangeTrustMaxLimit(t *testing.T) {
	kp0 := newKeypair0()
	txSourceAccount := makeTestAccount(kp0, "9605939170639898")

	changeTrust := ChangeTrust{
		Line: CreditAsset{"ABCD", kp0.Address()},
	}

	tx := Transaction{
		SourceAccount: &txSourceAccount,
		Operations:    []Operation{&changeTrust},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}
	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAiII0AAAAbAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAGAAAAAUFCQ0QAAAAA4Nxt4XJcrGZRYrUvrOc1sooiQ+QdEk1suS1wo+oucsV//////////wAAAAAAAAAB6i5yxQAAAEBen+Aa821qqfb+ASHhCg074ulPcCbIsUNvzUg2x92R/vRRKIGF5RztvPGBkktWkXLarDe5yqlBA+L3zpcVs+wB"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}
