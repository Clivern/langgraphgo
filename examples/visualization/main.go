package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a graph
	g := graph.NewStateGraph[map[string]any]()

	// 1. Define nodes
	g.AddNode("validate_input", "validate_input", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["valid"] = true
		return state, nil
	})

	g.AddNode("fetch_data", "fetch_data", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["data"] = "raw"
		return state, nil
	})

	g.AddNode("transform", "transform", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["data"] = "transformed"
		return state, nil
	})

	g.AddNode("enrich", "enrich", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["data"] = "enriched"
		return state, nil
	})

	g.AddNode("validate_output", "validate_output", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["output_valid"] = true
		return state, nil
	})

	g.AddNode("save", "save", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["saved"] = true
		return state, nil
	})

	g.AddNode("notify", "notify", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["notified"] = true
		return state, nil
	})

	// 2. Define edges (complex structure)
	g.SetEntryPoint("validate_input")
	g.AddEdge("validate_input", "fetch_data")
	g.AddEdge("fetch_data", "transform")
	g.AddEdge("transform", "enrich")
	g.AddEdge("enrich", "validate_output")
	g.AddEdge("validate_output", "save")
	g.AddEdge("save", "notify")
	g.AddEdge("notify", graph.END)

	// 3. Compile
	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	// 4. Visualize
	// Get the graph exporter
	exporter := graph.GetGraphForRunnable(runnable)

	fmt.Println("=== Mermaid Diagram ===")
	fmt.Println(exporter.DrawMermaid())

	fmt.Println("\n=== ASCII Diagram ===")
	fmt.Println(exporter.DrawASCII())
}
