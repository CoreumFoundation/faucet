package logger

import (
	"os"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/pflag"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

// re-export logger vars for convenience
var (
	ServiceDefaultConfig = logger.ServiceDefaultConfig
	ToolDefaultConfig    = logger.ToolDefaultConfig
	New                  = logger.New
	WithLogger           = logger.WithLogger
)

type (
	// Config re-export logger types for convenience
	Config = logger.Config
)

func newFlagRegister(fromFlagSet *pflag.FlagSet, exceptions ...string) func(*pflag.FlagSet) {
	return func(toFlagSet *pflag.FlagSet) {
		fromFlagSet.VisitAll(func(f *pflag.Flag) {
			if !lo.Contains(exceptions, f.Name) {
				toFlagSet.AddFlag(f)
			}
		})
	}
}

// ConfigureWithCLI configures logger based on CLI flags
func ConfigureWithCLI(defaultConfig logger.Config) (logger.Config, func(*pflag.FlagSet)) {
	flags := pflag.NewFlagSet("logger", pflag.ContinueOnError)
	flags.ParseErrorsWhitelist.UnknownFlags = true
	logger.AddFlags(defaultConfig, flags)
	// Dummy flag to turn off printing usage of this flag set
	flags.BoolP("help", "h", false, "")

	_ = flags.Parse(os.Args[1:])

	defaultConfig.Format = logger.Format(must.String(flags.GetString("log-format")))
	defaultConfig.Verbose = must.Bool(flags.GetBool("verbose"))
	if !validFormats[defaultConfig.Format] {
		panic(errors.Errorf("incorrect logging format %s", defaultConfig.Format))
	}

	return defaultConfig, newFlagRegister(flags, "help")
}

var validFormats = map[logger.Format]bool{
	logger.FormatConsole: true,
	logger.FormatJSON:    true,
	logger.FormatYAML:    true,
}
