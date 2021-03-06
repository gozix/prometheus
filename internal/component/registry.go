package component

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sarulabs/di/v2"
)

const (
	// DefRegistryName is a definition name.
	DefRegistryName = "prometheus.component.registry"
	// TagCollectorProvider indicates defs with prom collectors
	TagCollectorProvider = "prometheus.collector"
)

// DefRegistry is a prometheus registry definition getter.
func DefRegistry(registry *prometheus.Registry) di.Def {
	return di.Def{
		Name: DefRegistryName,
		Build: func(ctn di.Container) (interface{}, error) {
			var collectors []prometheus.Collector

			if registry == nil {
				registry = prometheus.NewRegistry()

				collectors = []prometheus.Collector{
					prometheus.NewGoCollector(),
					prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
				}
			}

			var err error
			for name, def := range ctn.Definitions() {
				for _, tag := range def.Tags {
					if tag.Name == TagCollectorProvider {
						var cs []prometheus.Collector
						if err = ctn.Fill(name, &cs); err != nil {
							return nil, err
						}
						collectors = append(collectors, cs...)
					}
				}
			}

			registry.MustRegister(collectors...)

			return registry, nil
		},
	}
}
