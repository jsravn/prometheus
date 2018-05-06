package main

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/gopherjs/gopherjs/js"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/tsdb"
)

type PromCache struct {
	storage     storage.Storage
	queryEngine *promql.Engine
	dir         string
	context     context.Context
}

func (p *PromCache) Close() {
	if p.storage != nil {
		p.storage.Close()
		os.RemoveAll(p.dir)
	}
}

func (p *PromCache) Init() {
	p.Close()
	var err error

	loglevel := promlog.AllowedLevel{}
	loglevel.Set("debug")
	logger := promlog.New(loglevel)
	level.Info(logger).Log("msg", "Starting Prometheus")

	p.dir, err = ioutil.TempDir("", "promCache")
	if err != nil {
		js.Debugger()
	}
	db, err := tsdb.Open(p.dir, nil, nil, &tsdb.Options{
		MinBlockDuration: model.Duration(24 * time.Hour),
		MaxBlockDuration: model.Duration(24 * time.Hour),
		NoLockfile:       true,
		WALFlushInterval: 1 * time.Hour,
	})
	if err != nil {
		js.Debugger()
	}
	p.storage = tsdb.Adapter(db, int64(0))
	p.queryEngine = promql.NewEngine(p.storage, nil)
	p.context = context.Background()
}

func main() {

	p := PromCache{}
	println("init:")
	p.Init()

	metric := labels.FromMap(map[string]string{
		"__name__": "ohai_js",
	})

	now := time.Now()
	samples := promql.Point{T: now.Unix() * 1000, V: 88.8}

	app, err := p.storage.Appender()
	if err != nil {
		js.Debugger()
	}
	app.Add(metric, samples.T, samples.V)
	app.Commit()

	println("query:")
	query, err := p.queryEngine.NewInstantQuery("ohai_js", now)
	if err != nil {
		js.Debugger()
	}
	println("query", query)
	println("query exec:")
	res := query.Exec(p.context)
	println("err", res.Err != nil)
	println("type", res.Value.Type())
	println("res", res)
	println("val -->", res.Value.String(), "<--")
	js.Debugger()
}
