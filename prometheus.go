package prometheus

import (
	gzviper "github.com/gozix/viper/v2"
	gzzap "github.com/gozix/zap/v2"
	"github.com/sarulabs/di/v2"

	"github.com/gozix/prometheus/internal/component"
)

// BundleName is default definition name.
const BundleName = "prometheus"

// DefRegistryName is a internal registry definition name
var DefRegistryName = component.DefRegistryName

// Bundle implements the glue.Bundle interface.
type Bundle struct{}

// NewBundle create bundle instance.
func NewBundle() *Bundle {
	return new(Bundle)
}

// Name is key implements the glue.Bundle interface.
func (b *Bundle) Name() string {
	return BundleName
}

// Build implements the glue.Bundle interface.
func (b *Bundle) Build(builder *di.Builder) error {
	return builder.Add(
		component.DefRegistry(),
		component.DefFlags(),
		component.DefServer(),
	)
}

// DependsOn implements the glue.DependsOn interface.
func (b *Bundle) DependsOn() []string {
	return []string{
		gzzap.BundleName,
		gzviper.BundleName,
	}
}
