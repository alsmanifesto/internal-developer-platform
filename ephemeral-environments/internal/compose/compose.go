// Package compose generates docker-compose configuration for ephemeral environments.
package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const composeTemplate = `version: "3.8"

services:
  {{ .EnvID }}:
    build:
      context: {{ .ProjectPath }}
    container_name: {{ .EnvID }}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.{{ .EnvID }}.rule=Host(` + "`" + `{{ .EnvID }}.local.ravon.dev` + "`" + `)"
      - "traefik.http.services.{{ .EnvID }}.loadbalancer.server.port=8080"
    networks:
      - traefik-net

  traefik:
    image: traefik:v2.11
    container_name: traefik
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
    networks:
      - traefik-net

networks:
  traefik-net:
    external: false
`

type composeData struct {
	EnvID       string
	ProjectPath string
}

// Generate writes a docker-compose.yml for the given environment into
// envs/<envID>/ (relative to CWD) and returns the directory path.
func Generate(envID, projectPath string) (string, error) {
	dir := filepath.Join("envs", envID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating env directory %q: %w", dir, err)
	}

	tmpl, err := template.New("compose").Parse(composeTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing compose template: %w", err)
	}

	composePath := filepath.Join(dir, "docker-compose.yml")
	f, err := os.Create(composePath)
	if err != nil {
		return "", fmt.Errorf("creating docker-compose.yml: %w", err)
	}
	defer f.Close()

	data := composeData{
		EnvID:       envID,
		ProjectPath: projectPath,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return "", fmt.Errorf("rendering compose template: %w", err)
	}

	return dir, nil
}
