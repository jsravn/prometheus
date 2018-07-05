package main

import (
	"encoding/json"
	"strconv"

	"github.com/cespare/xxhash"

	"github.com/gopherjs/gopherjs/js"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promcache"
	"github.com/prometheus/prometheus/promql"
)

/*
	this is the js <--> go interop layer, where inputs and outputs are js.Object
	and we do browser-esque things like
		js.Object -> json string -> json.Unmarshall to convert js to go
		errors are logged to console
		results are JSON.parse()'d objects
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
	p.SetMetricsAsString(str)
}

func (p *PromJS) SetMetricsAsString(rawJson string) {
	p.SetMetricsAsByteArray([]byte(rawJson))
}

func (p *PromJS) SetMetricsAsByteArray(rawJson []byte) {
	series := []promql.Series{}
	err := json.Unmarshal(rawJson, &series) // slow
	if err != nil {
		js.Global.Get("console").Call("error", "Load json error", err)
	}
	p.goLayer.SetMetrics(series)
}

// SetModel to declare the query for model rules
func (p *PromJS) SetModel(q string, tags map[string]string, mode string) {
	if mode == "" {
		mode = "main"
	}
	p.goLayer.SetModel(q, tags, mode)
}

// SetHealth to declare the query for health rules
func (p *PromJS) SetHealth(q string, tags map[string]string, mode string) {
	if mode == "" {
		mode = "main"
	}
	p.goLayer.SetHealth(q, tags, mode)
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

// FramedRangeQuery is RangeQuery plus start, end, step framing like the http api
// q is any promql that works in "Graph" tab of prometheus
// start, end are ints with unix timestamps: seconds since 1970
// step is seconds per bucket
func (p *PromJS) FramedRangeQuery(q string, start, end, step int) *js.Object {
	if start%step != 0 {
		js.Global.Get("console").Call("warn", "jittery RangeQuery:", q, "start:", start, "is misaligned with step:", step)
	}

	res, err := p.goLayer.FramedRangeQuery(q, start, end, step)
	if err != nil {
		js.Global.Get("console").Call("error", "RangeQuery error", err)
	}

	return response2json(res)
}

func (p *PromJS) Hash(m map[string]string) {

	if len(m) == 0 {
		m = map[string]string{
			"__name__":  "model",
			"name":      "ohai",
			"threshold": "0.1",
		}
	}

	ls := labels.FromMap(m)

	const sep = '\xff'
	b := make([]byte, 0, 1024)

	for _, v := range ls {
		b = append(b, v.Name...)
		b = append(b, sep)
		b = append(b, v.Value...)
		b = append(b, sep)
	}
	hash := js.Global.Get("XXH").Call("h64", string(b), 0)
	h64 := uint64(hash.Get("_a00").Uint64() +
		hash.Get("_a16").Uint64()<<16 +
		hash.Get("_a32").Uint64()<<32 +
		hash.Get("_a48").Uint64()<<48)

	nativeH := xxhash.Sum64(b)
	js.Global.Get("console").Call("log", hash.Call("toString", 16), strconv.FormatUint(h64, 16), strconv.FormatUint(nativeH, 16))
}
