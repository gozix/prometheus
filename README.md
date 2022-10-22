# GoZix Prometheus

[documentation-img]: https://img.shields.io/badge/godoc-reference-blue.svg?color=24B898&style=for-the-badge&logo=go&logoColor=ffffff
[documentation-url]: https://pkg.go.dev/github.com/gozix/prometheus/v2
[license-img]: https://img.shields.io/github/license/gozix/prometheus.svg?style=for-the-badge
[license-url]: https://github.com/gozix/prometheus/blob/master/LICENSE
[release-img]: https://img.shields.io/github/tag/gozix/prometheus.svg?label=release&color=24B898&logo=github&style=for-the-badge
[release-url]: https://github.com/gozix/prometheus/releases/latest
[build-status-img]: https://img.shields.io/github/actions/workflow/status/gozix/prometheus/go.yml?logo=github&style=for-the-badge
[build-status-url]: https://github.com/gozix/prometheus/actions
[go-report-img]: https://img.shields.io/badge/go%20report-A%2B-green?style=for-the-badge
[go-report-url]: https://goreportcard.com/report/github.com/gozix/prometheus
[code-coverage-img]: https://img.shields.io/codecov/c/github/gozix/prometheus.svg?style=for-the-badge&logo=codecov
[code-coverage-url]: https://codecov.io/gh/gozix/prometheus

[![License][license-img]][license-url]
[![Documentation][documentation-img]][documentation-url]

[![Release][release-img]][release-url]
[![Build Status][build-status-img]][build-status-url]
[![Go Report Card][go-report-img]][go-report-url]
[![Code Coverage][code-coverage-img]][code-coverage-url]

The bundle provide a Prometheus integration to GoZix application.

## Installation

```shell
go get github.com/gozix/prometheus/v2
```

## Dependencies

* [zap](https://github.com/gozix/prometheus)
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
--prometheus-port - Prometheus metrics port. The same `prometheus.port` config, but with high priority.
```

## Documentation

You can find documentation on [pkg.go.dev][documentation-url] and read source code if needed.

## Questions

If you have any questions, feel free to create an issue.
