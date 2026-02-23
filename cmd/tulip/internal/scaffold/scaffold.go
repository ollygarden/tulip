package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type templateData struct {
	Name       string
	Version    string
	OutputPath string // cleaned, e.g. "_build" not "./_build"
}

const configTemplate = `# Minimal configuration for {{.Name}} to start successfully.
# See https://opentelemetry.io/docs/collector/configuration/ for full reference.

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch: {}

exporters:
  debug: {}

service:
  pipelines:
    traces:
      receivers:  [otlp]
      processors: [batch]
      exporters:  [debug]
`

const dockerfileTemplate = `FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY manifest.yaml .

ARG OCB_VERSION={{.Version}}
RUN wget -qO /usr/bin/ocb \
    "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd%2Fbuilder%2Fv${OCB_VERSION}/ocb_${OCB_VERSION}_linux_amd64" && \
    chmod +x /usr/bin/ocb

RUN CGO_ENABLED=0 ocb --config=manifest.yaml

FROM alpine:latest

LABEL org.opencontainers.image.title="{{.Name}}"

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=builder /build/{{.OutputPath}}/{{.Name}} /{{.Name}}

ENTRYPOINT ["/{{.Name}}"]
CMD ["--config", "/etc/{{.Name}}/config.yaml"]
EXPOSE 4317 4318
`

// Write generates config.yaml and Dockerfile in distDir for the given distribution name.
func Write(distDir, name, version, outputPath string) error {
	data := templateData{
		Name:       name,
		Version:    version,
		OutputPath: strings.TrimPrefix(outputPath, "./"),
	}

	files := []struct {
		name     string
		tmplText string
	}{
		{"config.yaml", configTemplate},
		{"Dockerfile", dockerfileTemplate},
	}

	for _, f := range files {
		tmpl, err := template.New(f.name).Parse(f.tmplText)
		if err != nil {
			return fmt.Errorf("parsing %s template: %w", f.name, err)
		}

		path := filepath.Join(distDir, f.name)
		out, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("creating %s: %w", f.name, err)
		}

		if err := tmpl.Execute(out, data); err != nil {
			out.Close()
			return fmt.Errorf("executing %s template: %w", f.name, err)
		}

		if err := out.Close(); err != nil {
			return fmt.Errorf("closing %s: %w", f.name, err)
		}
	}

	return nil
}
