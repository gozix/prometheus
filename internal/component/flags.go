package component

import (
	gzglue "github.com/gozix/glue/v2"
	"github.com/sarulabs/di/v2"
	"github.com/spf13/pflag"
)

// DefFlagsName is a definition name.
const DefFlagsName = "prometheus.component.flags"

// portFlag is a value of mprometheus-port flag
var portFlagValue string

// DefFlags is a custom flags definition getter.
func DefFlags() di.Def {
	return di.Def{
		Name: DefFlagsName,
		Tags: []di.Tag{{
			Name: gzglue.TagRootPersistentFlags,
		}},
		Build: func(ctn di.Container) (interface{}, error) {
			var fs = pflag.FlagSet{}
			fs.StringVar(&portFlagValue, "prometheus-port", "", "prometheus metrics port")

			return &fs, nil
		},
	}
}
