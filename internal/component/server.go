package component

import (
	"context"
	"net"
	"net/http"
	"time"

	gzglue "github.com/gozix/glue/v2"
	gzviper "github.com/gozix/viper/v2"
	gzzap "github.com/gozix/zap/v2"
	"github.com/sarulabs/di/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DefServerName is a definition name.
const DefServerName = "prometheus.component.server"

// configPath is a first-level key for metrics config.
const configPath = "prometheus"

// Type checking.
var _ gzglue.PreRunner = (*preRunner)(nil)

type (
	// conf is a configuration struct.
	conf struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
		Path string `mapstructure:"path"`
	}

	// preRunner implements the glue.PreRunner
	preRunner struct {
		conf     conf
		registry *prometheus.Registry
		logger   *zap.Logger
		stop     func() error
	}
)

// DefServer is a http-server definition getter.
func DefServer() di.Def {
	return di.Def{
		Name: DefServerName,
		Tags: []di.Tag{{
			Name: gzglue.TagRootPersistentPreRunner,
		}},
		Build: func(ctn di.Container) (_ interface{}, err error) {
			var (
				log = ctn.Get(gzzap.BundleName).(*zap.Logger)
				cfg = ctn.Get(gzviper.BundleName).(*viper.Viper)
				reg = ctn.Get(DefRegistryName).(*prometheus.Registry)
			)

			var conf conf
			if err = cfg.UnmarshalKey(configPath, &conf); err != nil {
				return nil, err
			}

			if portFlagValue != "" {
				conf.Port = portFlagValue
			}

			if conf.Path == "" {
				conf.Path = "/"
			}

			return &preRunner{
				conf:     conf,
				registry: reg,
				logger:   log,
			}, nil
		},
		Close: func(obj interface{}) error {
			if runner, ok := obj.(*preRunner); ok {
				if runner.stop != nil {
					return runner.stop()
				}
			}
			return nil
		},
	}
}

// Run implements the glue.PreRunner.
func (r *preRunner) Run(_ context.Context) (err error) {
	var mux = http.NewServeMux()
	mux.Handle(r.conf.Path, promhttp.InstrumentMetricHandler(
		r.registry, promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{}),
	))

	var srv = &http.Server{
		Addr:    net.JoinHostPort(r.conf.Host, r.conf.Port),
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.logger.Panic("Error occurred during serve promhttp handler", zap.Error(err))
		}
	}()

	r.stop = func() error {
		var ctx, cancelFunc = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()
		return srv.Shutdown(ctx)
	}

	return nil
}
