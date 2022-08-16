package config

import (
	"os"
	"testing"

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

	assert.EqualValues(t, 12, port)
}

func TestWithEnv_WithoutPrefix(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	t.Setenv("PORT", "12")
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "pfx")
	require.NoError(t, err)

	assert.EqualValues(t, 1, port)
}

func TestWithEnv_OnlySetEnv(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	t.Setenv("PORT", "12")
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.EqualValues(t, 12, port)
}

func TestWithEnv_OnlySetFlag(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	os.Args = []string{"binary", "--port", "20"}
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.EqualValues(t, 20, port)
}

func TestWithEnv_FlagPrecedesEnv_SenEnvBeforeParse(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	os.Args = []string{"binary", "--port", "183"}
	t.Setenv("PORT", "182")
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.EqualValues(t, 183, port)
}

func TestWithEnv_FlagPrecedesEnv_SenEnvAfterParse(t *testing.T) {
	flagSet := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	var port int
	flagSet.IntVar(&port, "port", 1, "defines port")
	os.Args = []string{"binary", "--port", "183"}
	require.NoError(t, flagSet.Parse(os.Args[1:]))
	t.Setenv("PORT", "182")
	err := WithEnv(flagSet, "")
	require.NoError(t, err)

	assert.EqualValues(t, 183, port)
}
