# Tulip

Tulip is OllyGarden's commercially supported, vendor-agnostic distribution of the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/).

Send your telemetry data wherever you need it. Tulip ships a curated set of upstream OpenTelemetry components with enterprise support, quarterly releases, and long-term maintenance -- without locking you into any specific backend or vendor.

## How It Works

Tulip is built around a single file: [`distributions/tulip/manifest.yaml`](distributions/tulip/manifest.yaml).

This manifest defines exactly which OpenTelemetry components are included and supported. OllyGarden provides bug triage, CVE backports, and hands-on support for everything in this manifest.

**You can extend it.** Fork the manifest, add any component from the OpenTelemetry ecosystem, and build your own distribution. OllyGarden support covers the components in the official manifest; everything else is yours to manage.

## Quickstart

### Using the container image

```bash
docker pull cr.olly.garden/ollygarden/tulip/tulip:latest
docker run --rm -v $(pwd)/config.yaml:/etc/tulip/config.yaml cr.olly.garden/ollygarden/tulip/tulip:latest
```

Other image versions can be found in the [GitHub Container Registry](https://github.com/ollygarden/tulip/pkgs/container/tulip%2Ftulip).

### Building from source

```bash
make build
./distributions/tulip/_build/tulip --config=distributions/tulip/config.yaml
```

### Try it out

Send some test traces to a running Tulip instance:

```bash
go run github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen@latest \
  traces --otlp-endpoint localhost:4317 --otlp-insecure --rate 1 --duration 5s
```

## Included Components

| Category | Components |
|---|---|
| **Extensions** | zpages, healthcheck, pprof, basicauth, bearertokenauth, oauth2client, oidc |
| **Receivers** | otlp (gRPC + HTTP), nop |
| **Exporters** | otlp (gRPC), otlphttp, file, debug, nop |
| **Processors** | batch, attributes, resource, span, probabilisticsampler, filter, transform |
| **Connectors** | forward |

All components track the latest stable upstream OpenTelemetry Collector release. See [`manifest.yaml`](distributions/tulip/manifest.yaml) for exact versions.

## Extending Tulip

Want to add a component that isn't in the manifest?

**Build your own Tulip distribution.** Fork this repository, edit `distributions/tulip/manifest.yaml` to add your components, then build and validate:
   ```bash
   make generate       # regenerate sources from your updated manifest
   go mod tidy         # resolve new dependencies
   make docker-test    # build the Docker image and run the test suite against it
   ```

   Once the tests pass, your image is ready:
   ```bash
   make docker                          # builds tulip:local
   make docker DOCKER_IMAGE=my-tulip:1  # custom image name
   ```

## Installation

| Method | Details |
|---|---|
| **Container** | `docker pull cr.olly.garden/ollygarden/tulip/tulip:latest` |
| **System packages** | deb, rpm, apk available in [Releases](https://github.com/ollygarden/tulip/releases) |
| **Binary** | Download from [Releases](https://github.com/ollygarden/tulip/releases) (Linux/macOS, amd64/arm64) |
| **Build from source** | `make build` (requires Go 1.22+) |

### Configuration

- Default config: [`distributions/tulip/config.yaml`](distributions/tulip/config.yaml)
- system deployments: `/etc/tulip/config.yaml` and `/etc/tulip/tulip.conf`

## Releases

- **Quarterly stable releases** aligned with upstream OpenTelemetry Collector
- **LTS releases** every 18 months with extended maintenance
- **Version format:** `vYY.MM` (e.g., v25.11, v26.02)
- CVEs, performance regressions, and critical bugs are backported to all supported releases

For release notes and migration guides, see the [Releases](https://github.com/ollygarden/tulip/releases) page.

## Support

For support inquiries and pricing:
- Contact OllyGarden at ➡️ [ollygarden.com](https://ollygarden.com)
- Visit [ollygarden.com/tulip](https://ollygarden.com/tulip)

For upstream OpenTelemetry issues, refer to the [OpenTelemetry Collector repository](https://github.com/open-telemetry/opentelemetry-collector).


## License

Apache License, Version 2.0. See [LICENSE](LICENSE).
