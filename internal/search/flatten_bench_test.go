package search

import (
	"encoding/json"
	"testing"

	"storcat-wails/internal/fixture"
)

// BenchmarkLoadCatalogFlat40k measures LoadCatalogFlat against the default
// DCIM fixture shape (42,550 nodes), re-deriving the wire-size figure
// 23-RESEARCH.md's Pattern 1 measured rather than inheriting it (Pitfall
// N6). The fixture is generated once outside the timed loop; the returned
// FlatCatalog is marshalled once (also outside the loop) to report the
// megabyte figure the Wails bridge actually carries across per call.
func BenchmarkLoadCatalogFlat40k(b *testing.B) {
	dir := b.TempDir()
	path, nodeCount, _, err := fixture.WriteDCIMCatalog(dir, 50, 50, 16)
	if err != nil {
		b.Fatalf("WriteDCIMCatalog failed: %v", err)
	}

	s := NewService()

	flat, err := s.LoadCatalogFlat(path)
	if err != nil {
		b.Fatalf("LoadCatalogFlat failed: %v", err)
	}
	data, err := json.Marshal(flat)
	if err != nil {
		b.Fatalf("marshal FlatCatalog failed: %v", err)
	}
	megabytes := float64(len(data)) / (1024 * 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.LoadCatalogFlat(path); err != nil {
			b.Fatalf("LoadCatalogFlat failed: %v", err)
		}
	}
	b.StopTimer()

	b.ReportMetric(float64(nodeCount), "nodes")
	b.ReportMetric(megabytes, "MB")
}
