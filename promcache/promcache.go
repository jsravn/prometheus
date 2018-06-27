package promcache

import (
	"errors"
	"math"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

// maximum number of points returned in a query
const maxResolution = 300.

type keyValues map[string]string

// PromCache is an actor that keeps a temp Prometheus server
// and some caching for quick metric -> model -> health transforms
type PromCache struct {
	actor        Actor
	Server       *PromRunner
	err          error
	start        time.Time
	end          time.Time
	bucket       time.Duration
	needsRebuild bool
	metricCache  []promql.Series
	modelQuery   map[string]string
	modelLabels  map[string]keyValues
	healthQuery  map[string]string
	healthLabels map[string]keyValues
}

// NewPromCache initializes the prometheus backend and starts the actor
func NewPromCache() *PromCache {
	p := PromCache{
		Server:       NewPromRunner(),
		bucket:       time.Minute,
		modelQuery:   make(map[string]string),
		modelLabels:  make(map[string]keyValues),
		healthQuery:  make(map[string]string),
		healthLabels: make(map[string]keyValues),
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
		minutes := float64(newest-oldest) / 60.
		minutes = math.Ceil(minutes / maxResolution)
		p.bucket = time.Duration(minutes) * time.Minute
		p.metricCache = metrics
		p.needsRebuild = true
	})
}

func (p *PromCache) SetModel(q string, tags map[string]string, mode string) {
	p.actor.Tell(func() {
		p.modelQuery[mode] = q
		p.modelLabels[mode] = tags
		p.needsRebuild = true
	})
}

func (p *PromCache) SetHealth(q string, tags map[string]string, mode string) {
	p.actor.Tell(func() {
		p.healthQuery[mode] = q
		p.healthLabels[mode] = tags
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
	for mode := range p.modelQuery {
		p.eval(p.modelQuery[mode], p.modelLabels[mode])
	}
	for mode := range p.healthQuery {
		p.eval(p.healthQuery[mode], p.healthLabels[mode])
	}
	p.needsRebuild = false
}

func (p *PromCache) eval(q string, tags map[string]string) {
	res, err := p.Server.RangeQuery(q, p.start, p.end, p.bucket)
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
		res, err = p.Server.RangeQuery(q, p.start, p.end, p.bucket)
	})
	return res, err
}

func (p *PromCache) FramedRangeQuery(q string, start, end, step int) (res promql.Value, err error) {
	p.actor.Ask(func() {
		p.bucket = time.Duration(step) * time.Second
		p.rebuild()

		// clip to data range, to prevent lookups out of range
		dataStart := int(p.start.Unix())
		dataEnd := int(p.end.Unix())
		if start < dataStart {
			start = dataStart
		}
		if end < dataStart {
			end = dataStart
		}
		if start > dataEnd {
			start = dataEnd
		}
		if end > dataEnd {
			end = dataEnd
		}

		// align start on "step" to prevent jitter on small changes to start
		start = (start / step) * step

		res, err = p.Server.RangeQuery(
			q,
			time.Unix(int64(start), 0),
			time.Unix(int64(end), 0),
			time.Duration(step)*time.Second)
	})
	return res, err
}
