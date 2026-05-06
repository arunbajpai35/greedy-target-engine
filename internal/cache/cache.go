package cache

import (
	"context"
	"database/sql"
	"strings"
	"sync/atomic"

	"github.com/lib/pq"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

type rule struct {
	includeCountry, excludeCountry     map[string]struct{}
	includeOS, excludeOS               map[string]struct{}
	includeApp, excludeApp             map[string]struct{}
	hasIncCountry, hasIncOS, hasIncApp bool
}

type entry struct {
	c    models.Campaign
	rule rule
}

type snapshot struct {
	entries []entry
}

func (s *snapshot) match(app, country, os string) []models.Campaign {
	out := make([]models.Campaign, 0, 4)
	for _, e := range s.entries {
		r := e.rule
		if r.hasIncCountry {
			if _, ok := r.includeCountry[country]; !ok {
				continue
			}
		}
		if r.hasIncOS {
			if _, ok := r.includeOS[os]; !ok {
				continue
			}
		}
		if r.hasIncApp {
			if _, ok := r.includeApp[app]; !ok {
				continue
			}
		}
		if _, ok := r.excludeCountry[country]; ok {
			continue
		}
		if _, ok := r.excludeOS[os]; ok {
			continue
		}
		if _, ok := r.excludeApp[app]; ok {
			continue
		}
		out = append(out, e.c)
	}
	return out
}

type Cache struct {
	snap atomic.Pointer[snapshot]
}

func New() *Cache {
	c := &Cache{}
	c.snap.Store(&snapshot{})
	return c
}

func (c *Cache) Match(app, country, os string) []models.Campaign {
	return c.snap.Load().match(app, country, os)
}

func (c *Cache) Size() int {
	return len(c.snap.Load().entries)
}

func (c *Cache) Reload(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `
		SELECT c.cid, c.name, c.img, c.cta,
		       tr.include_country, tr.exclude_country,
		       tr.include_os, tr.exclude_os,
		       tr.include_app, tr.exclude_app
		FROM campaigns c
		JOIN targeting_rules tr ON c.cid = tr.cid
		WHERE c.status = 'ACTIVE'
		ORDER BY c.cid
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var entries []entry
	for rows.Next() {
		var (
			c                                  models.Campaign
			incC, excC, incO, excO, incA, excA []string
		)
		if err := rows.Scan(&c.ID, &c.Name, &c.Img, &c.CTA,
			pq.Array(&incC), pq.Array(&excC),
			pq.Array(&incO), pq.Array(&excO),
			pq.Array(&incA), pq.Array(&excA)); err != nil {
			return err
		}
		entries = append(entries, entry{
			c: c,
			rule: rule{
				includeCountry: toLowerSet(incC), hasIncCountry: incC != nil,
				excludeCountry: toLowerSet(excC),
				includeOS:      toLowerSet(incO), hasIncOS: incO != nil,
				excludeOS:      toLowerSet(excO),
				includeApp:     toLowerSet(incA), hasIncApp: incA != nil,
				excludeApp:     toLowerSet(excA),
			},
		})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	c.snap.Store(&snapshot{entries: entries})
	return nil
}

func toLowerSet(in []string) map[string]struct{} {
	if len(in) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(in))
	for _, v := range in {
		m[strings.ToLower(v)] = struct{}{}
	}
	return m
}
