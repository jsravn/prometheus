package promcache

import (
	"errors"
	"math"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

// PromCache is an actor that keeps a temp Prometheus server
// and some caching for quick metric -> model -> health transforms
type PromCache struct {
	actor        Actor
	Server       *PromRunner
	err          error
	start        time.Time
	end          time.Time
	needsRebuild bool
	metricCache  []promql.Series
	modelQuery   string
	modelLabels  map[string]string
	healthQuery  string
	healthLabels map[string]string
}

// NewPromCache initializes the prometheus backend and starts the actor
func NewPromCache() *PromCache {
	p := PromCache{
		Server: NewPromRunner(),
	}
	p.actor.Run()
	return &p
}

type queryResponse struct {
	Status string
	Data   struct {
		ResultType promql.ValueType
		Result     []promql.Series
	}
}

func (p *PromCache) Wait() {
	p.actor.Ask(func() {})
}

func (p *PromCache) LastError() error {
	return p.err
}

// SetMetrics caches a matrix of promql results
func (p *PromCache) SetMetrics(metrics []promql.Series) {
	p.actor.Tell(func() {
		// box everything on oldest and newest values in this result set
		oldest := int64(math.MaxInt64)
		newest := int64(math.MinInt64)
		for _, series := range metrics {
			numPts := len(series.Points)
			if numPts > 0 {
				ts := series.Points[0].T
				if ts < oldest {
					oldest = ts
				}
				ts = series.Points[numPts-1].T
				if ts > newest {
					newest = ts
				}
			}
		}
		oldest /= 1000 // milliseconds to seconds
		newest /= 1000
		oldest = (oldest / 60) * 60 // normalize to 60 sec buckets
		newest++                    // expand to include last data point

		p.start = time.Unix(oldest, 0)
		p.end = time.Unix(newest, 0)
		p.metricCache = metrics
		p.needsRebuild = true
	})
}

func (p *PromCache) SetModel(q string, tags map[string]string) {
	p.actor.Tell(func() {
		p.modelQuery = q
		p.modelLabels = tags
		p.needsRebuild = true
	})
}

func (p *PromCache) SetHealth(q string, tags map[string]string) {
	p.actor.Tell(func() {
		p.healthQuery = q
		p.healthLabels = tags
		p.needsRebuild = true
	})
}

func (p *PromCache) rebuild() {
	if !p.needsRebuild {
		return
	}
	p.Server.Reset()
	for _, series := range p.metricCache {
		p.Server.Load(series)
	}
	p.eval(p.modelQuery, p.modelLabels)
	p.eval(p.healthQuery, p.healthLabels)
	p.needsRebuild = false
}

func (p *PromCache) eval(q string, tags map[string]string) {
	res, err := p.Server.RangeQuery(q, p.start, p.end, time.Second*60)
	if err != nil {
		println(q + ": " + err.Error())
		p.err = err
		return
	}
	matrix, ok := res.(promql.Matrix)
	if !ok {
		p.err = errors.New("query result is not a range Vector")
		println(p.err.Error())
		return
	}
	for _, series := range matrix {
		// add tags
		orig := series.Metric.Map()
		for k, v := range tags {
			orig[k] = v
		}
		series.Metric = labels.FromMap(orig)
		// store
		err := p.Server.Load(series)
		if err != nil {
			println("load error: " + err.Error())
			p.err = err
			return
		}
	}
}

func (p *PromCache) InstantQuery(q string) (res promql.Value, err error) {
	p.actor.Ask(func() {
		p.rebuild()
		res, err = p.Server.InstantQuery(q, p.end)
	})
	return res, err
}

func (p *PromCache) RangeQuery(q string) (res promql.Value, err error) {
	p.actor.Ask(func() {
		p.rebuild()
		res, err = p.Server.RangeQuery(q, p.start, p.end, time.Minute)
	})
	return res, err
}
