package engine

import (
	"container/heap"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// RouteStep is a single conversion operation in a planned pipeline.
type RouteStep struct {
	Engine       string
	FromFormat   string
	ToFormat     string
	Cost         int
	FidelityNote string
	Warning      string
}

// ConversionRoute is the selected conversion pipeline metadata.
type ConversionRoute struct {
	Steps     []RouteStep
	TotalCost int
	Warnings  []string
}

func (r ConversionRoute) PrimaryEngine() string {
	if len(r.Steps) == 0 {
		return ""
	}
	eng := r.Steps[0].Engine
	switch eng {
	case "pandoc", "imagemagick", "ffmpeg", "data":
		return eng
	default:
		return "auto"
	}
}

func (r ConversionRoute) IsMultiHop() bool {
	return len(r.Steps) > 1
}

type plannerEdge struct {
	From         string
	To           string
	Engine       string
	Cost         int
	FidelityNote string
	Warning      string
}

var engineBaseCost = map[string]int{
	"data":        1,
	"pandoc":      2,
	"ffmpeg":      2,
	"imagemagick": 3,
	"pdftotext":   2,
}

var engineOrder = map[string]int{
	"data":        0,
	"pandoc":      1,
	"imagemagick": 2,
	"ffmpeg":      3,
	"pdftotext":   4,
}

func normalizeFormat(pathOrFormat string) string {
	if pathOrFormat == "" {
		return ""
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(pathOrFormat)), ".")
	if ext != "" {
		return canonicalFormat(ext)
	}
	return canonicalFormat(strings.TrimPrefix(strings.ToLower(pathOrFormat), "."))
}

func canonicalFormat(ext string) string {
	if ext == "yml" {
		return "yaml"
	}
	return ext
}

func buildPlannerEdges() []plannerEdge {
	edges := make([]plannerEdge, 0)
	for engineName, inputs := range engineInputFormats {
		outputs := engineOutputFormats[engineName]
		for from, canRead := range inputs {
			if !canRead {
				continue
			}
			from = canonicalFormat(from)
			for to, canWrite := range outputs {
				if !canWrite {
					continue
				}
				to = canonicalFormat(to)
				if from == to {
					continue
				}
				edges = append(edges, plannerEdge{
					From:         from,
					To:           to,
					Engine:       engineName,
					Cost:         engineBaseCost[engineName],
					FidelityNote: defaultFidelityNote(engineName),
				})
			}
		}
	}

	// Best-effort PDF input support via extracted text.
	edges = append(edges, plannerEdge{
		From:         "pdf",
		To:           "txt",
		Engine:       "pdftotext",
		Cost:         engineBaseCost["pdftotext"],
		FidelityNote: "best-effort text extraction",
		Warning:      "PDF input is handled via text extraction; layout and images may be lost",
	})

	return dedupeEdges(edges)
}

func defaultFidelityNote(engineName string) string {
	switch engineName {
	case "data":
		return "structured transformation"
	case "pandoc":
		return "document conversion"
	case "imagemagick":
		return "raster image conversion"
	case "ffmpeg":
		return "media transcode"
	default:
		return "conversion"
	}
}

func dedupeEdges(edges []plannerEdge) []plannerEdge {
	type key struct {
		from   string
		to     string
		engine string
	}
	seen := make(map[key]plannerEdge)
	for _, e := range edges {
		k := key{from: e.From, to: e.To, engine: e.Engine}
		if prev, ok := seen[k]; !ok || e.Cost < prev.Cost {
			seen[k] = e
		}
	}
	out := make([]plannerEdge, 0, len(seen))
	for _, e := range seen {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].To != out[j].To {
			return out[i].To < out[j].To
		}
		return out[i].Engine < out[j].Engine
	})
	return out
}

func PlanConversion(src, dst, forcedEngine string) (ConversionRoute, error) {
	srcFmt := normalizeFormat(src)
	dstFmt := normalizeFormat(dst)
	forcedEngine = strings.ToLower(strings.TrimSpace(forcedEngine))

	if srcFmt == "" || dstFmt == "" {
		return ConversionRoute{}, fmt.Errorf("cannot infer conversion format: %s -> %s", src, dst)
	}
	if srcFmt == dstFmt {
		return ConversionRoute{}, fmt.Errorf("source and target format are the same (%s)", srcFmt)
	}

	edges := buildPlannerEdges()
	if forcedEngine != "" && forcedEngine != "auto" {
		filtered := make([]plannerEdge, 0, len(edges))
		for _, e := range edges {
			if e.Engine == forcedEngine {
				filtered = append(filtered, e)
			}
		}
		edges = filtered
	}

	if len(edges) == 0 {
		if forcedEngine != "" && forcedEngine != "auto" {
			return ConversionRoute{}, fmt.Errorf("engine %s cannot perform this conversion route", forcedEngine)
		}
		return ConversionRoute{}, fmt.Errorf("no conversion path found")
	}

	// Direct conversions are always preferred when possible.
	if direct, ok := bestDirectRoute(edges, srcFmt, dstFmt); ok {
		return direct, nil
	}

	route, ok := shortestRoute(edges, srcFmt, dstFmt)
	if !ok {
		if forcedEngine != "" && forcedEngine != "auto" {
			return ConversionRoute{}, fmt.Errorf("engine %s cannot perform end-to-end conversion from %s to %s", forcedEngine, srcFmt, dstFmt)
		}
		return ConversionRoute{}, fmt.Errorf("no conversion path found from %s to %s", srcFmt, dstFmt)
	}
	return route, nil
}

func bestDirectRoute(edges []plannerEdge, srcFmt, dstFmt string) (ConversionRoute, bool) {
	candidates := make([]plannerEdge, 0)
	for _, e := range edges {
		if e.From == srcFmt && e.To == dstFmt {
			candidates = append(candidates, e)
		}
	}
	if len(candidates) == 0 {
		return ConversionRoute{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Cost != candidates[j].Cost {
			return candidates[i].Cost < candidates[j].Cost
		}
		return engineOrder[candidates[i].Engine] < engineOrder[candidates[j].Engine]
	})
	step := candidates[0]
	r := ConversionRoute{
		Steps: []RouteStep{{
			Engine:       step.Engine,
			FromFormat:   step.From,
			ToFormat:     step.To,
			Cost:         step.Cost,
			FidelityNote: step.FidelityNote,
			Warning:      step.Warning,
		}},
		TotalCost: step.Cost,
	}
	if step.Warning != "" {
		r.Warnings = append(r.Warnings, step.Warning)
	}
	return r, true
}

type nodeState struct {
	format string
	cost   int
	hops   int
	index  int
}

type minHeap []*nodeState

func (h minHeap) Len() int { return len(h) }

func (h minHeap) Less(i, j int) bool {
	if h[i].cost != h[j].cost {
		return h[i].cost < h[j].cost
	}
	if h[i].hops != h[j].hops {
		return h[i].hops < h[j].hops
	}
	return h[i].format < h[j].format
}

func (h minHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *minHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*nodeState)
	item.index = n
	*h = append(*h, item)
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

type prevStep struct {
	from string
	edge plannerEdge
}

func shortestRoute(edges []plannerEdge, srcFmt, dstFmt string) (ConversionRoute, bool) {
	adj := make(map[string][]plannerEdge)
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e)
	}
	for k := range adj {
		sort.Slice(adj[k], func(i, j int) bool {
			a := adj[k][i]
			b := adj[k][j]
			if a.Cost != b.Cost {
				return a.Cost < b.Cost
			}
			return engineOrder[a.Engine] < engineOrder[b.Engine]
		})
	}

	dist := map[string]int{srcFmt: 0}
	hops := map[string]int{srcFmt: 0}
	prev := map[string]prevStep{}

	pq := &minHeap{}
	heap.Init(pq)
	heap.Push(pq, &nodeState{format: srcFmt, cost: 0, hops: 0})

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*nodeState)
		if cur.format == dstFmt {
			break
		}
		if best, ok := dist[cur.format]; ok && cur.cost > best {
			continue
		}
		for _, e := range adj[cur.format] {
			nextCost := cur.cost + e.Cost
			nextHops := cur.hops + 1
			bestCost, seen := dist[e.To]
			bestHops := hops[e.To]
			if !seen || nextCost < bestCost || (nextCost == bestCost && nextHops < bestHops) {
				dist[e.To] = nextCost
				hops[e.To] = nextHops
				prev[e.To] = prevStep{from: cur.format, edge: e}
				heap.Push(pq, &nodeState{format: e.To, cost: nextCost, hops: nextHops})
			}
		}
	}

	if _, ok := dist[dstFmt]; !ok {
		return ConversionRoute{}, false
	}

	revSteps := make([]RouteStep, 0)
	warningsSet := make(map[string]bool)
	for cur := dstFmt; cur != srcFmt; {
		ps, ok := prev[cur]
		if !ok {
			return ConversionRoute{}, false
		}
		rs := RouteStep{
			Engine:       ps.edge.Engine,
			FromFormat:   ps.edge.From,
			ToFormat:     ps.edge.To,
			Cost:         ps.edge.Cost,
			FidelityNote: ps.edge.FidelityNote,
			Warning:      ps.edge.Warning,
		}
		revSteps = append(revSteps, rs)
		if rs.Warning != "" {
			warningsSet[rs.Warning] = true
		}
		cur = ps.from
	}

	steps := make([]RouteStep, 0, len(revSteps))
	for i := len(revSteps) - 1; i >= 0; i-- {
		steps = append(steps, revSteps[i])
	}
	warnings := make([]string, 0, len(warningsSet))
	for w := range warningsSet {
		warnings = append(warnings, w)
	}
	sort.Strings(warnings)

	return ConversionRoute{Steps: steps, TotalCost: dist[dstFmt], Warnings: warnings}, true
}
