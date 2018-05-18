package promcache

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/tsdb"
)

// PromRunner is a basic prometheus server with append storage and promql support
type PromRunner struct {
	storage     storage.Storage
	appender    storage.Appender
	needsCommit bool
	queryEngine *promql.Engine
	dir         string
	context     context.Context
	logger      log.Logger
}

// Close drops all the data
func (p *PromRunner) Close() {
	if p.storage != nil {
		p.storage.Close()
		os.RemoveAll(p.dir)
	}
}

// Reset drops all the data and starts over
func (p *PromRunner) Reset() {
	p.Close()
	p.init()
}

func (p *PromRunner) init() error {
	var err error

	p.dir, err = ioutil.TempDir("", "prom")
	if err != nil {
		return err
	}
	db, err := tsdb.Open(p.dir, p.logger, nil, &tsdb.Options{
		MinBlockDuration: model.Duration(24 * time.Hour),
		MaxBlockDuration: model.Duration(24 * time.Hour),
		NoLockfile:       true,
		WALFlushInterval: 1 * time.Hour,
	})
	if err != nil {
		return err
	}
	p.storage = tsdb.Adapter(db, int64(0))

	p.queryEngine = promql.NewEngine(p.storage, &promql.EngineOptions{
		MaxConcurrentQueries: 20,
		Timeout:              2 * time.Minute,
		Logger:               log.With(p.logger, "component", "query engine"),
	})
	p.context = context.Background()

	p.appender, err = p.storage.Appender()
	if err != nil {
		return err
	}
	return nil
}

// DebugPromRunner turns on Prometheus logging
func DebugPromRunner() *PromRunner {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowDebug())
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	level.Info(logger).Log("msg", "Starting Prometheus")
	p := &PromRunner{
		logger: logger,
	}
	p.init()
	return p
}

// NewPromRunner is quiet compared to DebugPromRunner
func NewPromRunner() *PromRunner {
	p := &PromRunner{
		logger: log.NewNopLogger(),
	}
	p.init()
	return p
}

// Load a raw Series
func (p *PromRunner) Load(series promql.Series) (err error) {
	refid := uint64(0)
	for _, pt := range series.Points {
		if refid == 0 {
			refid, err = p.appender.Add(series.Metric, pt.T, pt.V)
		} else {
			err = p.appender.AddFast(series.Metric, refid, pt.T, pt.V)
		}
		if err != nil {
			return err
		}
	}
	p.needsCommit = true
	return nil
}

// InstantQuery promql at a given momemt in time
func (p *PromRunner) InstantQuery(q string, t time.Time) (promql.Value, error) {
	if p.needsCommit {
		p.needsCommit = false
		p.appender.Commit()
	}

	query, err := p.queryEngine.NewInstantQuery(q, t)
	if err != nil {
		return nil, err
	}

	res := query.Exec(p.context)
	return res.Value, res.Err
}

// RangeQuery promql from here to there, bucketed on "interval"
func (p *PromRunner) RangeQuery(q string, start, end time.Time, interval time.Duration) (promql.Value, error) {
	if p.needsCommit {
		p.needsCommit = false
		p.appender.Commit()
	}

	query, err := p.queryEngine.NewRangeQuery(q, start, end, interval)
	if err != nil {
		return nil, err
	}

	res := query.Exec(p.context)
	return res.Value, res.Err
}
