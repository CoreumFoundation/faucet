package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAddress(t *testing.T) {
	testCases := []struct {
		name           string
		address        string
		expectedPrefix string
		verifyError    bool
	}{
		{
			name:           "correct devcore",
			address:        "devcore10krrrqxxy948n5p9xvwgq6krgy9hg5g8svaz62",
			expectedPrefix: "devcore",
			verifyError:    false,
		},
		{
			name:           "correct cosmos",
			address:        "cosmos169ltjnyvfcxhfxa03xc6qdsu9068ceynym2awg",
			expectedPrefix: "cosmos",
			verifyError:    false,
		},
		{
			name:           "correct with different private key type",
			address:        "testtoken10r5hnadz9vj3lqjfachadxgwww9jpvwu7z067chwdn47mnka895q5q8lrk",
			expectedPrefix: "testtoken",
			verifyError:    false,
		},
		{
			name:           "checksum failing",
			address:        "invalid10krrrqxxy948n5p9xvwgq6krgy9hg5g8svaz62",
			expectedPrefix: "",
			verifyError:    true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
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
