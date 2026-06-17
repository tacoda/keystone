package primitive

import (
	"fmt"
	"sort"
	"strings"
)

// GraphFormat picks the renderer. Both formats produce text the
// dashboard or a downstream tool can consume — Mermaid is rendered
// inline in browsers via the htmx page; DOT pipes into graphviz.
type GraphFormat string

const (
	GraphMermaid GraphFormat = "mermaid"
	GraphDOT     GraphFormat = "dot"
)

// RenderGraph builds the primitive-relationship graph and serializes
// it in the requested format. Edges:
//
//   deps:   <p> --deps--> <target>
//   traces: <p> --traces--> <corpus/target>
//
// Nodes are kept compact: `<kind>/<id>` is the canonical name; the
// `kind` value also drives the CSS / Mermaid class so the dashboard
// renders each layer in its accent color.
func RenderGraph(primitives []Primitive, format GraphFormat) string {
	nodes, edges := buildEdges(primitives)
	switch format {
	case GraphDOT:
		return renderDOT(nodes, edges)
	default:
		return renderMermaid(nodes, edges)
	}
}

type graphEdge struct {
	from, to, via string
}

func buildEdges(primitives []Primitive) ([]Primitive, []graphEdge) {
	byKey := map[string]Primitive{}
	for _, p := range primitives {
		byKey[p.Kind+"/"+p.ID] = p
		byKey[p.ID] = p
	}
	var edges []graphEdge
	for _, p := range primitives {
		src := p.Kind + "/" + p.ID
		for _, d := range p.Deps {
			edges = append(edges, graphEdge{from: src, to: d, via: "deps"})
		}
		for _, t := range p.Traces {
			target := t
			if !strings.Contains(t, "/") {
				target = "corpus/" + t
			}
			if _, ok := byKey[target]; !ok {
				if cp, ok := byKey["corpus/"+t]; ok {
					target = cp.Kind + "/" + cp.ID
				}
			}
			edges = append(edges, graphEdge{from: src, to: target, via: "traces"})
		}
	}
	// Stable sorts.
	out := make([]Primitive, len(primitives))
	copy(out, primitives)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].ID < out[j].ID
	})
	sort.SliceStable(edges, func(i, j int) bool {
		if edges[i].from != edges[j].from {
			return edges[i].from < edges[j].from
		}
		return edges[i].to < edges[j].to
	})
	return out, edges
}

func nodeName(key string) string {
	// Mermaid IDs can't contain `/` or `:` — replace with `_`.
	r := strings.NewReplacer("/", "_", ":", "_", "-", "_", ".", "_")
	return r.Replace(key)
}

func renderMermaid(nodes []Primitive, edges []graphEdge) string {
	var b strings.Builder
	fmt.Fprintln(&b, "flowchart LR")
	for _, p := range nodes {
		key := p.Kind + "/" + p.ID
		fmt.Fprintf(&b, "  %s[\"<small>%s</small><br/>%s\"]:::%s\n", nodeName(key), p.Kind, p.ID, p.Kind)
	}
	for _, e := range edges {
		label := ""
		if e.via == "deps" {
			label = "deps"
		} else {
			label = "traces"
		}
		fmt.Fprintf(&b, "  %s -- %s --> %s\n", nodeName(e.from), label, nodeName(e.to))
	}
	// Per-kind colors aligned with the dashboard palette.
	for _, line := range []string{
		"classDef guide    fill:#13161d,stroke:#4ade9a,color:#4ade9a",
		"classDef rule     fill:#13161d,stroke:#4ade9a,color:#4ade9a",
		"classDef corpus   fill:#13161d,stroke:#60a5fa,color:#60a5fa",
		"classDef sensor   fill:#13161d,stroke:#f59e0b,color:#f59e0b",
		"classDef action   fill:#13161d,stroke:#e879a0,color:#e879a0",
		"classDef playbook fill:#13161d,stroke:#e879a0,color:#e879a0",
		"classDef skill    fill:#13161d,stroke:#a78bfa,color:#a78bfa",
		"classDef subagent fill:#13161d,stroke:#a78bfa,color:#a78bfa",
		"classDef command  fill:#13161d,stroke:#a78bfa,color:#a78bfa",
		"classDef persona  fill:#13161d,stroke:#a78bfa,color:#a78bfa",
		"classDef eval     fill:#13161d,stroke:#f59e0b,color:#f59e0b",
		"classDef source   fill:#13161d,stroke:#60a5fa,color:#60a5fa",
	} {
		fmt.Fprintln(&b, "  "+line)
	}
	return b.String()
}

func renderDOT(nodes []Primitive, edges []graphEdge) string {
	var b strings.Builder
	fmt.Fprintln(&b, "digraph keystone {")
	fmt.Fprintln(&b, `  bgcolor="#0d1018"; node [shape=box, style=filled, fillcolor="#13161d", fontcolor="#e7e9ee", color="#2a2f3a"]; edge [color="#6b7280", fontcolor="#9aa0a6"];`)
	for _, p := range nodes {
		key := p.Kind + "/" + p.ID
		fmt.Fprintf(&b, "  %s [label=%q];\n", nodeName(key), p.Kind+"\n"+p.ID)
	}
	for _, e := range edges {
		fmt.Fprintf(&b, "  %s -> %s [label=%q];\n", nodeName(e.from), nodeName(e.to), e.via)
	}
	fmt.Fprintln(&b, "}")
	return b.String()
}
