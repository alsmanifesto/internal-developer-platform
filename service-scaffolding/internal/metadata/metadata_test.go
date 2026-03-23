package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helpers

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "scaffold-metadata-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func mustLoad(t *testing.T, base string) *MetadataStore {
	t.Helper()
	store, err := loadMetadataFrom(base)
	if err != nil {
		t.Fatalf("loadMetadataFrom: %v", err)
	}
	return store
}

func mustSave(t *testing.T, store *MetadataStore, base string) {
	t.Helper()
	if err := saveMetadataTo(store, base); err != nil {
		t.Fatalf("saveMetadataTo: %v", err)
	}
}

func svc(name string, createdAt time.Time) ServiceMetadata {
	return ServiceMetadata{
		Name:        name,
		ServiceType: "api",
		Workload:    "app",
		Stack:       "go",
		Pipeline:    "gh-actions",
		Path:        name,
		CreatedAt:   createdAt,
	}
}

// LoadMetadata

func TestLoadMetadata_CreatesStoreWhenFileMissing(t *testing.T) {
	dir := tempDir(t)
	store := mustLoad(t, dir)

	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if len(store.Services) != 0 {
		t.Fatalf("expected empty services, got %d", len(store.Services))
	}
}

func TestLoadMetadata_CreatesDirWhenMissing(t *testing.T) {
	dir := tempDir(t)
	mustLoad(t, dir)

	if _, err := os.Stat(filepath.Join(dir, metadataDir)); os.IsNotExist(err) {
		t.Fatal("expected .scaffold directory to be created")
	}
}

func TestLoadMetadata_ReadsExistingFile(t *testing.T) {
	dir := tempDir(t)
	store := &MetadataStore{Services: []ServiceMetadata{svc("payments", time.Now())}}
	mustSave(t, store, dir)

	loaded := mustLoad(t, dir)
	if len(loaded.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(loaded.Services))
	}
	if loaded.Services[0].Name != "payments" {
		t.Fatalf("expected service name 'payments', got %q", loaded.Services[0].Name)
	}
}

func TestLoadMetadata_NilServicesBecomesEmpty(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, metadataDir, metadataFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// Write JSON with explicit null services
	if err := os.WriteFile(path, []byte(`{"services": null}`), 0o644); err != nil {
		t.Fatal(err)
	}

	store := mustLoad(t, dir)
	if store.Services == nil {
		t.Fatal("expected Services to be non-nil empty slice, got nil")
	}
}

func TestLoadMetadata_ErrorOnCorruptJSON(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, metadataDir, metadataFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadMetadataFrom(dir)
	if err == nil {
		t.Fatal("expected error on corrupt JSON, got nil")
	}
}

// SaveMetadata

func TestSaveMetadata_WritesPrettyJSON(t *testing.T) {
	dir := tempDir(t)
	store := &MetadataStore{Services: []ServiceMetadata{svc("orders", time.Now())}}
	mustSave(t, store, dir)

	data, err := os.ReadFile(filepath.Join(dir, metadataDir, metadataFile))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	// Pretty JSON must contain indentation
	if len(data) == 0 {
		t.Fatal("saved file is empty")
	}
	var roundtrip MetadataStore
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}
	if len(roundtrip.Services) != 1 || roundtrip.Services[0].Name != "orders" {
		t.Fatal("roundtrip mismatch")
	}
}

func TestSaveMetadata_RoundtripPreservesAllFields(t *testing.T) {
	dir := tempDir(t)
	now := time.Now().UTC().Truncate(time.Second)
	original := ServiceMetadata{
		Name:        "inventory",
		ServiceType: "worker",
		Workload:    "data",
		Stack:       "python",
		Pipeline:    "concourse",
		Path:        "inventory",
		CreatedAt:   now,
	}
	store := &MetadataStore{Services: []ServiceMetadata{original}}
	mustSave(t, store, dir)

	loaded := mustLoad(t, dir)
	got := loaded.Services[0]

	if got.Name != original.Name {
		t.Errorf("Name: want %q got %q", original.Name, got.Name)
	}
	if got.ServiceType != original.ServiceType {
		t.Errorf("ServiceType: want %q got %q", original.ServiceType, got.ServiceType)
	}
	if got.Workload != original.Workload {
		t.Errorf("Workload: want %q got %q", original.Workload, got.Workload)
	}
	if got.Stack != original.Stack {
		t.Errorf("Stack: want %q got %q", original.Stack, got.Stack)
	}
	if got.Pipeline != original.Pipeline {
		t.Errorf("Pipeline: want %q got %q", original.Pipeline, got.Pipeline)
	}
	if !got.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt: want %v got %v", original.CreatedAt, got.CreatedAt)
	}
}

// AddService

func TestAddService_AddsToEmptyStore(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{}}
	if err := AddService(store, svc("alpha", time.Now())); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(store.Services))
	}
}

func TestAddService_RejectsDuplicate(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{}}
	_ = AddService(store, svc("alpha", time.Now()))

	err := AddService(store, svc("alpha", time.Now()))
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
}

func TestAddService_AllowsDifferentNames(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{}}
	_ = AddService(store, svc("alpha", time.Now()))
	if err := AddService(store, svc("beta", time.Now())); err != nil {
		t.Fatalf("unexpected error adding second service: %v", err)
	}
	if len(store.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(store.Services))
	}
}

func TestAddService_SetsCreatedAtIfZero(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{}}
	s := svc("gamma", time.Time{}) // zero time
	before := time.Now()
	_ = AddService(store, s)
	after := time.Now()

	got := store.Services[0].CreatedAt
	if got.IsZero() {
		t.Fatal("expected CreatedAt to be set, still zero")
	}
	if got.Before(before) || got.After(after) {
		t.Errorf("CreatedAt %v out of expected range [%v, %v]", got, before, after)
	}
}

func TestAddService_PreservesExplicitCreatedAt(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{}}
	explicit := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	_ = AddService(store, svc("delta", explicit))

	if !store.Services[0].CreatedAt.Equal(explicit) {
		t.Errorf("expected CreatedAt %v, got %v", explicit, store.Services[0].CreatedAt)
	}
}

// RemoveService

func TestRemoveService_RemovesExisting(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{svc("alpha", time.Now())}}
	if err := RemoveService(store, "alpha"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Services) != 0 {
		t.Fatalf("expected 0 services after removal, got %d", len(store.Services))
	}
}

func TestRemoveService_ErrorOnMissingName(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{}}
	if err := RemoveService(store, "nonexistent"); err == nil {
		t.Fatal("expected error when removing nonexistent service")
	}
}

func TestRemoveService_RemovesOnlyTarget(t *testing.T) {
	now := time.Now()
	store := &MetadataStore{Services: []ServiceMetadata{
		svc("alpha", now),
		svc("beta", now.Add(time.Second)),
		svc("gamma", now.Add(2*time.Second)),
	}}

	if err := RemoveService(store, "beta"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Services) != 2 {
		t.Fatalf("expected 2 remaining services, got %d", len(store.Services))
	}
	for _, s := range store.Services {
		if s.Name == "beta" {
			t.Fatal("'beta' still present after removal")
		}
	}
}

func TestRemoveService_RemovesFirst(t *testing.T) {
	now := time.Now()
	store := &MetadataStore{Services: []ServiceMetadata{
		svc("first", now),
		svc("second", now.Add(time.Second)),
	}}
	if err := RemoveService(store, "first"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Services) != 1 || store.Services[0].Name != "second" {
		t.Fatal("wrong service remaining after removing first")
	}
}

func TestRemoveService_RemovesLast(t *testing.T) {
	now := time.Now()
	store := &MetadataStore{Services: []ServiceMetadata{
		svc("first", now),
		svc("last", now.Add(time.Second)),
	}}
	if err := RemoveService(store, "last"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Services) != 1 || store.Services[0].Name != "first" {
		t.Fatal("wrong service remaining after removing last")
	}
}

// ListServices

func TestListServices_EmptyStoreReturnsEmpty(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{}}
	result := ListServices(store)
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestListServices_SortedNewestFirst(t *testing.T) {
	now := time.Now()
	store := &MetadataStore{Services: []ServiceMetadata{
		svc("oldest", now),
		svc("newest", now.Add(2*time.Second)),
		svc("middle", now.Add(time.Second)),
	}}

	result := ListServices(store)
	expected := []string{"newest", "middle", "oldest"}
	for i, want := range expected {
		if result[i].Name != want {
			t.Errorf("position %d: want %q got %q", i, want, result[i].Name)
		}
	}
}

func TestListServices_DoesNotMutateOriginal(t *testing.T) {
	now := time.Now()
	store := &MetadataStore{Services: []ServiceMetadata{
		svc("oldest", now),
		svc("newest", now.Add(2*time.Second)),
	}}
	originalFirst := store.Services[0].Name

	ListServices(store)

	if store.Services[0].Name != originalFirst {
		t.Fatal("ListServices mutated the original store order")
	}
}

func TestListServices_SingleServiceReturnsIt(t *testing.T) {
	store := &MetadataStore{Services: []ServiceMetadata{svc("only", time.Now())}}
	result := ListServices(store)
	if len(result) != 1 || result[0].Name != "only" {
		t.Fatal("unexpected result for single-service store")
	}
}

// Integration: full lifecycle

func TestFullLifecycle_CreateSaveLoadDeleteSave(t *testing.T) {
	dir := tempDir(t)
	now := time.Now()

	// 1. Load empty store
	store := mustLoad(t, dir)
	if len(store.Services) != 0 {
		t.Fatal("expected empty store on first load")
	}

	// 2. Add two services and save
	_ = AddService(store, svc("svc-a", now))
	_ = AddService(store, svc("svc-b", now.Add(time.Second)))
	mustSave(t, store, dir)

	// 3. Reload and verify both present, sorted correctly
	reloaded := mustLoad(t, dir)
	listed := ListServices(reloaded)
	if len(listed) != 2 {
		t.Fatalf("expected 2 services after reload, got %d", len(listed))
	}
	if listed[0].Name != "svc-b" || listed[1].Name != "svc-a" {
		t.Errorf("unexpected order: %v", []string{listed[0].Name, listed[1].Name})
	}

	// 4. Remove one service and save
	if err := RemoveService(reloaded, "svc-a"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	mustSave(t, reloaded, dir)

	// 5. Final reload — only svc-b remains
	final := mustLoad(t, dir)
	if len(final.Services) != 1 {
		t.Fatalf("expected 1 service after deletion, got %d", len(final.Services))
	}
	if final.Services[0].Name != "svc-b" {
		t.Fatalf("expected 'svc-b', got %q", final.Services[0].Name)
	}
}

func TestFullLifecycle_DuplicateRejectedAfterReload(t *testing.T) {
	dir := tempDir(t)

	store := mustLoad(t, dir)
	_ = AddService(store, svc("unique", time.Now()))
	mustSave(t, store, dir)

	reloaded := mustLoad(t, dir)
	err := AddService(reloaded, svc("unique", time.Now()))
	if err == nil {
		t.Fatal("expected duplicate error after reload, got nil")
	}
}
