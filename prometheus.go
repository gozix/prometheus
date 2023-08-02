// Copyright 2020 Alexander Gromov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package prometheus

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gozix/di"
	"github.com/gozix/glue/v3"
	gzViper "github.com/gozix/viper/v3"
	gzZap "github.com/gozix/zap/v3"

	"github.com/prometheus/client_golang/prometheus"
	prometheusCollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	// BundleName is default definition name.
	BundleName = "prometheus"

	// TagCollectorProvider is tag marks prometheus collectors.
	TagCollectorProvider = "prometheus.collector"
)

type (
	// Bundle implements the glue.Bundle interface.
	Bundle struct {
		registry      *prometheus.Registry
		flagPortValue string
	}

	// Option interface.
	Option interface {
		apply(b *Bundle)
	}

	// optionFunc wraps a func so it satisfies the Option interface.
	optionFunc func(b *Bundle)
)

// Bundle implements glue.Bundle interface.
var _ glue.Bundle = (*Bundle)(nil)

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

func (b *Bundle) Name() string {
	return BundleName
}

func (b *Bundle) Build(builder di.Builder) error {
	return builder.Apply(
		di.Provide(
			b.provideRegistry,
			di.Constraint(0, di.Optional(true), di.WithTags(TagCollectorProvider)),
			di.As(new(prometheus.Gatherer)),
			di.As(new(prometheus.Registerer)),
		),
		di.Provide(b.provideFlagSet, glue.AsPersistentFlags()),
		di.Provide(b.providePreRunner, glue.AsPersistentPreRunner()),
	)
}

func (b *Bundle) DependsOn() []string {
	return []string{
		gzZap.BundleName,
		gzViper.BundleName,
	}
}

func (b *Bundle) provideFlagSet() *pflag.FlagSet {
	var flagSet = pflag.NewFlagSet(BundleName, pflag.ExitOnError)
	flagSet.StringVar(&b.flagPortValue, "prometheus-port", "", "prometheus metrics port")

	return flagSet
}

func (b *Bundle) provideRegistry(collectors []prometheus.Collector) (_ *prometheus.Registry, err error) {
	var registry = b.registry
	if registry == nil {
		registry = prometheus.NewRegistry()
		collectors = []prometheus.Collector{
			prometheusCollectors.NewGoCollector(),
			prometheusCollectors.NewProcessCollector(prometheusCollectors.ProcessCollectorOpts{}),
		}
	}

	for _, collector := range collectors {
		if err = registry.Register(collector); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

func (b *Bundle) providePreRunner(
	cfg *viper.Viper,
	logger *zap.Logger,
	registry *prometheus.Registry,
) (_ glue.PreRunner, _ func() error, err error) {
	// use this hack for UnmarshalKey
	// see https://github.com/spf13/viper/issues/188
	var configPath = BundleName
	var cfgPath = cfg.Sub(configPath)
	if cfgPath != nil {
		for _, key := range cfg.Sub(configPath).AllKeys() {
			key = configPath + "." + key
			cfg.Set(key, cfg.Get(key))
		}
	}

	var conf = struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
		Path string `mapstructure:"path"`
	}{}

	if err = cfg.UnmarshalKey(configPath, &conf); err != nil {
		return nil, nil, err
	}

	if b.flagPortValue != "" {
		conf.Port = b.flagPortValue
	}

	if conf.Path == "" {
		conf.Path = "/"
	}

	var mux = http.NewServeMux()
	mux.Handle(conf.Path, promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	))

	var srv = &http.Server{
		Addr:    net.JoinHostPort(conf.Host, conf.Port),
		Handler: mux,
	}

	var preRunner = glue.PreRunnerFunc(func(ctx context.Context) error {
		go func() {
			if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Panic("Error occurred during serve prometheus http handler", zap.Error(err))
			}
		}()

		return nil
	})

	var closer = func() error {
		var ctx, cancelFunc = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		return srv.Shutdown(ctx)
	}

	return preRunner, closer, nil
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
