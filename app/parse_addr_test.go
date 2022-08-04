package app

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {
	// pk,_ :=types.GenerateSecp256k1Key()
	// pk
	prv := cosmossecp256k1.GenPrivKey()
	// println(prv.PubKey().Address().String())
	public, _, _ := ed25519.GenerateKey(rand.Reader)
	println(sdk.AccAddress(prv.PubKey().Address()).String())
	sdk.GetConfig().SetBech32PrefixForAccount("testtoken", "testtoken")
	println(sdk.AccAddress(public).String())
	// sdk.AccAddressFromBech32()
	t.Fail()
}

func TestParseAddress(t *testing.T) {
	testCases := []struct {
		address        string
		expectedPrefix string
		verifyError    bool
	}{
		{
			address:        "devcore10krrrqxxy948n5p9xvwgq6krgy9hg5g8svaz62",
			expectedPrefix: "devcore",
			verifyError:    false,
		},
		{
			address:        "cosmos169ltjnyvfcxhfxa03xc6qdsu9068ceynym2awg",
			expectedPrefix: "cosmos",
			verifyError:    false,
		},
		{
			address:        "testtoken10r5hnadz9vj3lqjfachadxgwww9jpvwu7z067chwdn47mnka895q5q8lrk",
			expectedPrefix: "testtoken",
			verifyError:    false,
		},
		{
			address:        "invalid10krrrqxxy948n5p9xvwgq6krgy9hg5g8svaz62",
			expectedPrefix: "",
			verifyError:    true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run("", func(t *testing.T) {
			assertT := assert.New(t)
			prefix, addr, err := parseAddress(tc.address)
			assertT.EqualValues(tc.expectedPrefix, prefix)
			if !tc.verifyError {
				assertT.NoError(err)
				assertT.NotNil(addr)
			} else {
				assertT.Error(err)
				assertT.Nil(addr)
			}
		})
	}
}
