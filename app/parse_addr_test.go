package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
