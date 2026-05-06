package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

func set(vs ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(vs))
	for _, v := range vs {
		m[v] = struct{}{}
	}
	return m
}

func TestSnapshotMatch(t *testing.T) {
	snap := &snapshot{entries: []entry{
		{
			c: models.Campaign{ID: "spotify"},
			rule: rule{
				includeCountry: set("us", "canada"), hasIncCountry: true,
			},
		},
		{
			c: models.Campaign{ID: "duolingo"},
			rule: rule{
				excludeCountry: set("us"),
				includeOS:      set("android", "ios"), hasIncOS: true,
			},
		},
		{
			c: models.Campaign{ID: "subwaysurfer"},
			rule: rule{
				includeOS:  set("android"), hasIncOS: true,
				includeApp: set("com.gametion.ludokinggame"), hasIncApp: true,
			},
		},
	}}

	cases := []struct {
		name           string
		app, ctry, ros string
		want           []string
	}{
		{"us+android+ludo", "com.gametion.ludokinggame", "us", "android", []string{"spotify", "subwaysurfer"}},
		{"germany+android", "com.test", "germany", "android", []string{"duolingo"}},
		{"us+android+other-app", "com.test", "us", "android", []string{"spotify"}},
		{"web everywhere", "com.test", "zz", "web", nil},
		{"exclude wins", "com.test", "us", "android", []string{"spotify"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := snap.match(c.app, c.ctry, c.ros)
			ids := make([]string, 0, len(got))
			for _, g := range got {
				ids = append(ids, g.ID)
			}
			assert.ElementsMatch(t, c.want, ids)
		})
	}
}
