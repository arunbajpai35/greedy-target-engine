package cache

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

func buildSnapshot(n int) *snapshot {
	rng := rand.New(rand.NewSource(1))
	countries := []string{"us", "canada", "germany", "india", "japan", "brazil", "france"}
	oses := []string{"android", "ios", "web"}

	entries := make([]entry, 0, n)
	for i := 0; i < n; i++ {
		entries = append(entries, entry{
			c: models.Campaign{ID: fmt.Sprintf("c-%d", i), Name: "x", Img: "x", CTA: "Go"},
			rule: rule{
				includeCountry: set(countries[rng.Intn(len(countries))]),
				hasIncCountry:  true,
				includeOS:      set(oses[rng.Intn(len(oses))]),
				hasIncOS:       true,
			},
		})
	}
	return &snapshot{entries: entries}
}

func benchMatch(b *testing.B, n int) {
	snap := buildSnapshot(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = snap.match("com.example", "us", "android")
	}
}

func BenchmarkMatch_100(b *testing.B)    { benchMatch(b, 100) }
func BenchmarkMatch_1000(b *testing.B)   { benchMatch(b, 1000) }
func BenchmarkMatch_10000(b *testing.B)  { benchMatch(b, 10000) }
func BenchmarkMatch_100000(b *testing.B) { benchMatch(b, 100000) }
