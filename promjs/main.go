package main

import (
	"encoding/json"
	"time"

	"github.com/gopherjs/gopherjs/js"
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
	goLayer *PromCache
}

// New eg:  p = promCache.New()
func New() *js.Object {
	return js.MakeWrapper(&PromJS{NewPromCache()})
}

// Close to shut down the prometheus engine, Load and Query will now fail
func (p *PromJS) Close() {
	p.goLayer.Close()
}

// Reset to clear database and start over, ready for Load()
func (p *PromJS) Reset() {
	p.goLayer.Reset()
}

// Load json in format of a prometheus response:  data.result[0]
func (p *PromJS) Load(o *js.Object) {
	str := js.Global.Get("JSON").Call("stringify", o).String()
	series := promql.Series{}
	err := json.Unmarshal([]byte(str), &series)
	if err != nil {
		js.Global.Get("console").Call("error", "Load json error", err)
	}
	err = p.goLayer.Load(series)
	if err != nil {
		js.Global.Get("console").Call("error", "Load error", err)
	}
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

// InstantQuery eg: InstantQuery("metric", 1525786621.627)
// q is any promql that works in "Console" tab of prometheus
// time is in unix seconds, we ignore the fractional milliseconds
func (p *PromJS) InstantQuery(q string, unixTime float64) *js.Object {
	t := time.Unix(int64(unixTime), 0)
	res, err := p.goLayer.InstantQuery(q, t)
	if err != nil {
		js.Global.Get("console").Call("error", "InstantQuery error", err)
	}
	return response2json(res)
}

// RangeQuery eg: RangeQuery("metric", 1525780621.123, 1525786621.627, 10)
// q is any promql that works in "Graph" tab of prometheus
// times are unix seconds, ignoring the fractional milliseconds
// interval is seconds
func (p *PromJS) RangeQuery(q string, start, end float64, interval int) *js.Object {
	res, err := p.goLayer.RangeQuery(
		q,
		time.Unix(int64(start), 0),
		time.Unix(int64(end), 0),
		time.Duration(interval)*time.Second)
	if err != nil {
		js.Global.Get("console").Call("error", "RangeQuery error", err)
	}
	return response2json(res)
}
