# GoZix Prometheus

## Dependencies

* [zap](https://github.com/gozix/zap)
* [viper](https://github.com/gozix/viper)

## Configuration example

```json
{
  "prometheus": {
    "host": "0.0.0.0",
    "port": 8683,
    "path": "/metrics"
  }
}
```

## Flags

```
--metrics-port - Prometheus metrics port. The same `prometheus.port` config, but with high priority.
```
