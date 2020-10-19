package prometheus

import (
	gzsql "github.com/gozix/sql/v2"
	gzviper "github.com/gozix/viper/v2"
	gzzap "github.com/gozix/zap/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sarulabs/di/v2"

	"github.com/gozix/prometheus/internal/component"
)

const (
	// BundleName is default definition name.
	BundleName = "prometheus"
	// // TagCollectorProvider is alias of component.TagCollectorProvider.
	TagCollectorProvider = component.TagCollectorProvider
)

// DefRegistryName is a internal registry definition name
var DefRegistryName = component.DefRegistryName

type (
	// Bundle implements the glue.Bundle interface.
	Bundle struct {
		registry *prometheus.Registry
	}

	// Option interface.
	Option interface {
		apply(b *Bundle)
	}

	// optionFunc wraps a func so it satisfies the Option interface.
	optionFunc func(b *Bundle)
)

// Registry option makes it possible to set a custom registry.
func Registry(registry *prometheus.Registry) Option {
	return optionFunc(func(b *Bundle) {
		b.registry = registry
	})
}

// NewBundle create bundle instance.
func NewBundle(options ...Option) *Bundle {
	var b = new(Bundle)
	for _, option := range options {
		option.apply(b)
	}
	return b
}

// Name is key implements the glue.Bundle interface.
func (b *Bundle) Name() string {
	return BundleName
}

// Build implements the glue.Bundle interface.
func (b *Bundle) Build(builder *di.Builder) error {
	return builder.Add(
		component.DefRegistry(b.registry),
		component.DefFlags(),
		component.DefServer(),
	)
}

// DependsOn implements the glue.DependsOn interface.
func (b *Bundle) DependsOn() []string {
	return []string{
		gzzap.BundleName,
		gzviper.BundleName,
		gzsql.BundleName,
	}
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
