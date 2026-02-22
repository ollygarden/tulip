package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type templateData struct {
	Name string
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

const dockerfileTemplate = `FROM alpine:latest

LABEL org.opencontainers.image.title="{{.Name}}"

COPY --chmod=755 {{.Name}} /{{.Name}}
COPY config.yaml /etc/{{.Name}}/config.yaml

EXPOSE 4317

ENTRYPOINT ["/{{.Name}}"]
CMD ["--config", "/etc/{{.Name}}/config.yaml"]
`

// Write generates config.yaml and Dockerfile in distDir for the given distribution name.
func Write(distDir string, name string) error {
	data := templateData{Name: name}

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
