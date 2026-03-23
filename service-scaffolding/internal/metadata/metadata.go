// Package metadata manages the persistent service registry for scaffold.
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	metadataDir  = ".scaffold"
	metadataFile = "services.json"
)

// ServiceMetadata holds the configuration for a single scaffolded service.
type ServiceMetadata struct {
	Name        string    `json:"name"`
	ServiceType string    `json:"service_type"`
	Workload    string    `json:"workload"`
	Stack       string    `json:"stack"`
	Pipeline    string    `json:"pipeline"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
}

// MetadataStore is the root structure persisted to disk.
type MetadataStore struct {
	Services []ServiceMetadata `json:"services"`
}

func metadataPath() string {
	return metadataPathIn(".")
}

func metadataPathIn(base string) string {
	return filepath.Join(base, metadataDir, metadataFile)
}

// LoadMetadata reads the metadata file from disk, creating it if it does not exist.
func LoadMetadata() (*MetadataStore, error) {
	return loadMetadataFrom(".")
}

func loadMetadataFrom(base string) (*MetadataStore, error) {
	dir := filepath.Join(base, metadataDir)
	path := filepath.Join(dir, metadataFile)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create metadata dir: %w", err)
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &MetadataStore{Services: []ServiceMetadata{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read metadata file: %w", err)
	}

	var store MetadataStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parse metadata file: %w", err)
	}

	if store.Services == nil {
		store.Services = []ServiceMetadata{}
	}

	return &store, nil
}

// SaveMetadata writes the metadata store to disk as pretty-printed JSON.
func SaveMetadata(store *MetadataStore) error {
	return saveMetadataTo(store, ".")
}

func saveMetadataTo(store *MetadataStore, base string) error {
	dir := filepath.Join(base, metadataDir)
	path := filepath.Join(dir, metadataFile)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create metadata dir: %w", err)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write metadata file: %w", err)
	}

	return nil
}

// AddService appends a service to the store, preventing duplicate names.
func AddService(store *MetadataStore, svc ServiceMetadata) error {
	for _, s := range store.Services {
		if s.Name == svc.Name {
			return fmt.Errorf("service %q already exists", svc.Name)
		}
	}

	if svc.CreatedAt.IsZero() {
		svc.CreatedAt = time.Now()
	}

	store.Services = append(store.Services, svc)
	return nil
}

// RemoveService removes a service by name. Returns an error if not found.
func RemoveService(store *MetadataStore, name string) error {
	for i, s := range store.Services {
		if s.Name == name {
			store.Services = append(store.Services[:i], store.Services[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("service %q not found", name)
}

// ListServices returns services sorted by creation time, newest first.
func ListServices(store *MetadataStore) []ServiceMetadata {
	sorted := make([]ServiceMetadata, len(store.Services))
	copy(sorted, store.Services)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	return sorted
}
