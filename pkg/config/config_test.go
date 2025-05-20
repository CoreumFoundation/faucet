package config

import (
	"math"
	"os"
	"testing"

	"github.com/samber/lo"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithEnv_WithPrefix(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	t.Setenv("PFX_PORT", "12")
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "pfx")
	require.NoError(t, err)

	assert.Equal(t, 12, port)
}

func TestWithEnv_WithoutPrefix(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	t.Setenv("PORT", "12")
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "some_prefix")
	require.NoError(t, err)

	assert.Equal(t, 1, port)
}

func TestWithEnv_OnlySetEnv(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	t.Setenv("PORT", "12")
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.Equal(t, 12, port)
}

func TestWithEnv_OnlySetFlag(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	revert := setArgAndRevert([]string{"binary", "--port", "20"})
	defer revert()
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.Equal(t, 20, port)
}

func TestWithEnv_FlagPrecedesEnv_SenEnvBeforeParse(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	os.Args = []string{"binary", "--port", "183"}
	revert := setArgAndRevert([]string{"binary", "--port", "183"})
	defer revert()
	t.Setenv("PORT", "182")
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.Equal(t, 183, port)
}

func TestWithEnv_FlagPrecedesEnv_SenEnvAfterParse(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	revert := setArgAndRevert([]string{"binary", "--port", "183"})
	defer revert()
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	t.Setenv("PORT", "182")
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.Equal(t, 183, port)
}

func setArgAndRevert(args []string) func() {
	oldArg := lo.Subset(os.Args, -2, math.MaxUint64)
	os.Args = args
	return func() {
		os.Args = oldArg
	}
}
