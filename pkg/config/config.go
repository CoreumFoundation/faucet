package config

import (
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/spf13/pflag"
)

// NewModuleManager returns new module manager
func NewModuleManager() module.BasicManager {
	return module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
	)
}

// WithEnv gets a flagSet and sets its values, with values read from env vars.
// This function should be called only after all the flags are defined.
func WithEnv(f *pflag.FlagSet, prefix string) error {
	var flagNames []string
	f.VisitAll(func(flag *pflag.Flag) {
		flagNames = append(flagNames, flag.Name)
	})

	for _, fn := range flagNames {
		flag := f.Lookup(fn)
		if flag.Changed {
			continue
		}

		name := flag.Name
		if prefix != "" {
			name = prefix + "_" + name
		}
		name = strings.ReplaceAll(strings.ToUpper(name), "-", "_")
		envValue := os.Getenv(name)
		if envValue != "" {
			flag.DefValue = envValue
			if err := flag.Value.Set(envValue); err != nil {
				return err
			}
		}
	}

	return nil
}
