package component

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sarulabs/di/v2"
)

// DefRegistryName is a definition name.
const DefRegistryName = "prometheus.component.registry"

// DefRegistry is a prometheus registry definition getter.
func DefRegistry() di.Def {
	return di.Def{
		Name: DefRegistryName,
		Build: func(ctn di.Container) (interface{}, error) {
			var registry = prometheus.NewRegistry()

			registry.MustRegister(prometheus.NewGoCollector())
			registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

			return registry, nil
		},
	}
}
