// Copyright 2020 Alexander Gromov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package prometheus

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
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

	// tagPrometheusFlagSet is tag marks bundle flag set.
	tagPrometheusFlagSet = "prometheus.flag_set"

	// flagPrometheusPort is flag name.
	flagPrometheusPort = "prometheus-port"
)

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
		di.Provide(b.provideFlagSet, glue.AsPersistentFlags(), di.Tags{{Name: tagPrometheusFlagSet}}),
		di.Provide(
			b.providePreRunner,
			glue.AsPersistentPreRunner(),
			di.Constraint(2, di.WithTags(tagPrometheusFlagSet)),
		),
	)
}

func (b *Bundle) DependsOn() []string {
	return []string{
		gzZap.BundleName,
		gzViper.BundleName,
	}
}

func (b *Bundle) provideFlagSet() (*pflag.FlagSet, error) {
	var flagSet = pflag.NewFlagSet(BundleName, pflag.ContinueOnError)
	flagSet.String(flagPrometheusPort, "", "prometheus metrics port")
	flagSet.ParseErrorsWhitelist.UnknownFlags = true

	var err = flagSet.Parse(os.Args)
	if errors.Is(err, pflag.ErrHelp) {
		err = nil
	}

	return flagSet, err
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
	flagSet *pflag.FlagSet,
	registry *prometheus.Registry,
) (glue.PreRunner, func() error, error) {
	// use this hack for UnmarshalKey
	// see https://github.com/spf13/viper/issues/188
	if cfgPath := cfg.Sub(BundleName); cfgPath != nil {
		for _, key := range cfg.Sub(BundleName).AllKeys() {
			key = BundleName + "." + key
			cfg.Set(key, cfg.Get(key))
		}
	}

	var conf = struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
		Path string `mapstructure:"path"`
	}{}

	if err := cfg.UnmarshalKey(BundleName, &conf); err != nil {
		return nil, nil, err
	}

	if flag := flagSet.Lookup(flagPrometheusPort); flag != nil && flag.Value.String() != "" {
		conf.Port = flag.Value.String()
	}

	if conf.Path == "" {
		conf.Path = "/"
	}

	var mux = http.NewServeMux()
	mux.Handle(conf.Path, promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	))

	var (
		srv = &http.Server{
			Addr:    net.JoinHostPort(conf.Host, conf.Port),
			Handler: mux,
		}
		log = logger.With(zap.String("addr", srv.Addr))
	)

	var preRunner = glue.PreRunnerFunc(func(ctx context.Context) error {
		log.Info("Starting prometheus HTTP server")

		go func() {
			var err = srv.ListenAndServe()
			switch {
			case errors.Is(err, http.ErrServerClosed):
				log.Info("Gracefully shutting down the HTTP server")

			case err != nil:
				log.Error("Error occurred during serve prometheus http server", zap.Error(err))
			}
		}()

		log.Info("Prometheus HTTP server started")

		return nil
	})

	var closer = func() error {
		var timeout = 10 * time.Second
		log.Info("Stopping prometheus HTTP server", zap.Duration("timeout", timeout))

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
