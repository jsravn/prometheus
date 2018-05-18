package main

import (
	"encoding/json"

	"github.com/gopherjs/gopherjs/js"
	"github.com/prometheus/prometheus/promcache"
	"github.com/prometheus/prometheus/promql"
)

/*
	this is the js -> go interop layer, where inputs and outputs are js.Object
	and we do silly things like js.Object -> json string -> json.Unmarshall to convert js to go
*/

func main() {
	js.Global.Set("promCache", map[string]interface{}{
		"New": New,
	})
}

// PromJS is PromCache wrapped as a JS object
type PromJS struct {
	goLayer *promcache.PromCache
}

// New eg:  p = promCache.New()
func New() *js.Object {
	return js.MakeWrapper(&PromJS{promcache.NewPromCache()})
}

// SetMetrics to pass in a promql matrix
func (p *PromJS) SetMetrics(o *js.Object) {
	str := js.Global.Get("JSON").Call("stringify", o).String()
	series := []promql.Series{}
	err := json.Unmarshal([]byte(str), &series)
	if err != nil {
		js.Global.Get("console").Call("error", "Load json error", err)
	}
	p.goLayer.SetMetrics(series)
}

// SetModel to declare the query for model rules
func (p *PromJS) SetModel(q string, tags map[string]string) {
	p.goLayer.SetModel(q, tags)
}

// SetHealth to declare the query for health rules
func (p *PromJS) SetHealth(q string, tags map[string]string) {
	p.goLayer.SetHealth(q, tags)
}

func response2json(res promql.Value) *js.Object {
	b, err := json.Marshal(&res)
	if err != nil {
		js.Global.Get("console").Call("error", "Query json error", err)
	}
	s := string(b)
	if len(s) == 0 {
		s = "[]"
	}
	return js.Global.Get("JSON").Call("parse", s)
}

// InstantQuery eg: InstantQuery("metric")
// q is any promql that works in "Console" tab of prometheus
func (p *PromJS) InstantQuery(q string) *js.Object {
	res, err := p.goLayer.InstantQuery(q)
	if err != nil {
		js.Global.Get("console").Call("error", "InstantQuery error", err)
	}
	return response2json(res)
}

// RangeQuery eg: RangeQuery("metric")
// q is any promql that works in "Graph" tab of prometheus
func (p *PromJS) RangeQuery(q string) *js.Object {
	res, err := p.goLayer.RangeQuery(q)
	if err != nil {
		js.Global.Get("console").Call("error", "RangeQuery error", err)
	}
	return response2json(res)
}
