package store

import (
	"testing"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/stretchr/testify/assert"
)

func TestNewFalkorDBGraph(t *testing.T) {
	t.Run("Valid connection string with custom graph name", func(t *testing.T) {
		g, err := NewFalkorDBGraph("falkordb://localhost:6379/custom_graph")
		assert.NoError(t, err)
		assert.NotNil(t, g)
		fg := g.(*FalkorDBGraph)
		assert.Equal(t, "custom_graph", fg.graphName)
		assert.NotNil(t, fg.client)
		fg.Close()
	})

	t.Run("Valid connection string with default graph name", func(t *testing.T) {
		g, err := NewFalkorDBGraph("falkordb://localhost:6379")
		assert.NoError(t, err)
		assert.NotNil(t, g)
		fg := g.(*FalkorDBGraph)
		assert.Equal(t, "rag", fg.graphName) // Default graph name
		fg.Close()
	})

	t.Run("Invalid URL", func(t *testing.T) {
		g, err := NewFalkorDBGraph("://invalid")
		assert.Error(t, err)
		assert.Nil(t, g)
		assert.Contains(t, err.Error(), "invalid connection string")
	})

	t.Run("Missing host", func(t *testing.T) {
		g, err := NewFalkorDBGraph("falkordb:///graph")
		assert.Error(t, err)
		assert.Nil(t, g)
		assert.Contains(t, err.Error(), "missing host")
	})

	t.Run("NewKnowledgeGraph factory", func(t *testing.T) {
		g, err := NewKnowledgeGraph("falkordb://localhost:6379/graph")
		if err == nil {
			assert.NotNil(t, g)
			if fg, ok := g.(*FalkorDBGraph); ok {
				fg.Close()
			}
		}
	})
}

func TestSanitizeLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple label", "Person", "Person"},
		{"Label with space", "Person Age", "Person_Age"},
		{"Label with special chars", "Person-Type@123", "Person_Type_123"},
		{"Empty label", "", "Entity"},
		{"Only special chars", "@#$%", "____"}, // Special chars become underscores, not Entity
		{"Mixed case", "MyEntity", "MyEntity"},
		{"Numbers", "Entity123", "Entity123"},
		{"Underscores preserved", "My_Entity", "My_Entity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeLabel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPropsToString(t *testing.T) {
	t.Run("String properties", func(t *testing.T) {
		props := map[string]any{"name": "test", "age": 30}
		s := propsToString(props)
		assert.Contains(t, s, "name")
		assert.Contains(t, s, "age")
		assert.Contains(t, s, "{")
		assert.Contains(t, s, "}")
	})

	t.Run("Float32 slice embedding", func(t *testing.T) {
		props := map[string]any{
			"name":      "entity",
			"embedding": []float32{0.1, 0.2, 0.3},
		}
		s := propsToString(props)
		assert.Contains(t, s, "embedding")
		assert.Contains(t, s, "[")
		assert.Contains(t, s, "]")
	})

	t.Run("Boolean and numeric values", func(t *testing.T) {
		props := map[string]any{
			"active": true,
			"count":  42,
			"ratio":  3.14,
		}
		s := propsToString(props)
		assert.Contains(t, s, "active")
		assert.Contains(t, s, "count")
		assert.Contains(t, s, "ratio")
	})

	t.Run("Empty map", func(t *testing.T) {
		props := map[string]any{}
		s := propsToString(props)
		assert.Equal(t, "{}", s)
	})
}

func TestEntityToMap(t *testing.T) {
	t.Run("Entity with all fields", func(t *testing.T) {
		e := &rag.Entity{
			ID:         "1",
			Name:       "John",
			Type:       "Person",
			Embedding:  []float32{0.1, 0.2},
			Properties: map[string]any{"age": 30},
		}
		m := entityToMap(e)
		assert.Equal(t, "John", m["name"])
		assert.Equal(t, "Person", m["type"])
		assert.Equal(t, 30, m["age"])
		assert.NotNil(t, m["embedding"])
	})

	t.Run("Entity without embedding", func(t *testing.T) {
		e := &rag.Entity{
			ID:         "2",
			Name:       "Jane",
			Type:       "Person",
			Properties: map[string]any{"city": "NYC"},
		}
		m := entityToMap(e)
		assert.Equal(t, "Jane", m["name"])
		assert.Equal(t, "Person", m["type"])
		assert.Equal(t, "NYC", m["city"])
		assert.Nil(t, m["embedding"])
	})

	t.Run("Entity with empty properties", func(t *testing.T) {
		e := &rag.Entity{
			ID:         "3",
			Name:       "Test",
			Type:       "Type",
			Properties: map[string]any{},
		}
		m := entityToMap(e)
		assert.Equal(t, "Test", m["name"])
		assert.Equal(t, "Type", m["type"])
	})
}

func TestRelationshipToMap(t *testing.T) {
	t.Run("Relationship with all fields", func(t *testing.T) {
		r := &rag.Relationship{
			ID:         "1",
			Source:     "s",
			Target:     "t",
			Type:       "KNOWS",
			Weight:     0.8,
			Confidence: 0.9,
			Properties: map[string]any{"since": 2020},
		}
		m := relationshipToMap(r)
		assert.Equal(t, "KNOWS", m["type"])
		assert.Equal(t, 0.8, m["weight"])
		assert.Equal(t, 0.9, m["confidence"])
		assert.Equal(t, 2020, m["since"])
	})

	t.Run("Relationship with empty properties", func(t *testing.T) {
		r := &rag.Relationship{
			ID:         "2",
			Source:     "a",
			Target:     "b",
			Type:       "RELATED",
			Properties: map[string]any{},
		}
		m := relationshipToMap(r)
		assert.Equal(t, "RELATED", m["type"])
		assert.Contains(t, m, "weight")
		assert.Contains(t, m, "confidence")
	})
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"String input", "hello", "hello"},
		{"Byte slice", []byte("world"), "world"},
		{"Integer", 123, "123"},
		{"Float", 3.14, "3.14"},
		{"Boolean", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseNode(t *testing.T) {
	t.Run("Standard format with labels and properties", func(t *testing.T) {
		// Format: [id, [labels], [[key1, val1], [key2, val2]]]
		obj := []any{
			int64(1),
			[]any{[]byte("Person")},
			[]any{
				[]any{int64(1), int64(2), "id"},
				[]any{int64(1), int64(4), "test"},
				[]any{int64(2), int64(4), "name"},
				[]any{int64(2), int64(4), "John"},
			},
		}
		e := parseNode(obj)
		assert.NotNil(t, e)
		assert.Equal(t, "Person", e.Type)
		assert.Equal(t, "test", e.ID)
		assert.Equal(t, "John", e.Name)
	})

	t.Run("KV format", func(t *testing.T) {
		obj := []any{
			[]any{"id", "node1"},
			[]any{"labels", []any{"Person"}},
			[]any{"properties", []any{
				[]any{"name", "Alice"},
				[]any{"id", "alice1"},
			}},
		}
		e := parseNode(obj)
		assert.NotNil(t, e)
		assert.Equal(t, "alice1", e.ID)
		assert.Equal(t, "Alice", e.Name)
		assert.Equal(t, "Person", e.Type)
	})

	t.Run("Invalid format", func(t *testing.T) {
		e := parseNode("not a slice")
		assert.Nil(t, e)
	})

	t.Run("Empty slice", func(t *testing.T) {
		e := parseNode([]any{})
		assert.NotNil(t, e)
	})

	t.Run("Complex nested structure", func(t *testing.T) {
		obj := []any{
			int64(1),
			[]any{
				int64(2),
				[]any{[]byte("Label")},
				[]any{
					[]any{int64(1), int64(2), "id"},
					[]any{int64(1), int64(3), "id1"},
				},
			},
		}
		e := parseNode(obj)
		assert.NotNil(t, e)
		assert.Equal(t, "id1", e.ID)
	})
}

func TestParseNodeKV(t *testing.T) {
	t.Run("Complete KV pairs", func(t *testing.T) {
		pairs := []any{
			[]any{"id", "entity1"},
			[]any{"labels", []any{"Person"}},
			[]any{"properties", []any{
				[]any{"name", "Bob"},
				[]any{"type", "User"},
				[]any{"age", "30"},
			}},
		}
		e := parseNodeKV(pairs)
		assert.NotNil(t, e)
		assert.Equal(t, "entity1", e.ID)
		assert.Equal(t, "User", e.Type)
		assert.Equal(t, "Bob", e.Name)
		assert.Equal(t, "30", e.Properties["age"])
	})

	t.Run("Invalid pairs", func(t *testing.T) {
		pairs := []any{
			"not a pair",
			[]any{"single"},
		}
		e := parseNodeKV(pairs)
		assert.NotNil(t, e)
		assert.Empty(t, e.ID)
	})
}

func TestParseEdge(t *testing.T) {
	t.Run("Standard edge format", func(t *testing.T) {
		obj := []any{
			int64(1),
			[]byte("KNOWS"),
			int64(2),
			int64(3),
			[]any{
				[]any{"id", "rel1"},
				[]any{"weight", 0.5},
			},
		}
		rel := parseEdge(obj, "source1", "target1")
		assert.NotNil(t, rel)
		assert.Equal(t, "source1", rel.Source)
		assert.Equal(t, "target1", rel.Target)
		assert.Equal(t, "KNOWS", rel.Type)
		assert.Equal(t, "rel1", rel.ID)
	})

	t.Run("KV edge format", func(t *testing.T) {
		obj := []any{
			[]any{"id", "edge1"},
			[]any{"type", "RELATED"},
			[]any{"properties", []any{
				[]any{"id", "edge_id1"},
				[]any{"strength", "high"},
			}},
		}
		rel := parseEdge(obj, "src", "dst")
		assert.NotNil(t, rel)
		assert.Equal(t, "src", rel.Source)
		assert.Equal(t, "dst", rel.Target)
		assert.Equal(t, "RELATED", rel.Type)
		assert.Equal(t, "edge_id1", rel.ID)
		assert.Equal(t, "high", rel.Properties["strength"])
	})

	t.Run("Invalid format", func(t *testing.T) {
		rel := parseEdge("not a slice", "src", "dst")
		assert.Nil(t, rel)
	})

	t.Run("Short slice", func(t *testing.T) {
		obj := []any{int64(1), []byte("TYPE")}
		rel := parseEdge(obj, "s", "t")
		assert.Nil(t, rel)
	})

	t.Run("String type", func(t *testing.T) {
		obj := []any{
			int64(1),
			"WORKS_WITH",
			int64(2),
			int64(3),
			[]any{},
		}
		rel := parseEdge(obj, "a", "b")
		assert.NotNil(t, rel)
		assert.Equal(t, "WORKS_WITH", rel.Type)
	})
}

func TestParseFalkorDBProperties(t *testing.T) {
	t.Run("Even number of properties", func(t *testing.T) {
		props := []any{
			[]any{int64(1), int64(2), "name"},
			[]any{int64(2), int64(4), "John"},
			[]any{int64(3), int64(4), "type"},
			[]any{int64(4), int64(6), "Person"},
		}
		e := &rag.Entity{Properties: make(map[string]any)}
		parseFalkorDBProperties(props, e)
		assert.Equal(t, "John", e.Name)
		assert.Equal(t, "Person", e.Type)
	})

	t.Run("Odd number of properties", func(t *testing.T) {
		props := []any{
			[]any{int64(1), int64(2), "id"},
			[]any{int64(2), int64(5), "test1"},
			[]any{int64(3), int64(3), "age"},
		}
		e := &rag.Entity{Properties: make(map[string]any)}
		parseFalkorDBProperties(props, e)
		assert.Equal(t, "test1", e.ID)
	})

	t.Run("Custom properties", func(t *testing.T) {
		props := []any{
			[]any{int64(1), int64(3), "city"},
			[]any{int64(2), int64(3), "NYC"},
			[]any{int64(3), int64(3), "country"},
			[]any{int64(4), int64(3), "USA"},
		}
		e := &rag.Entity{Properties: make(map[string]any)}
		parseFalkorDBProperties(props, e)
		assert.Equal(t, "NYC", e.Properties["city"])
		assert.Equal(t, "USA", e.Properties["country"])
	})
}

func TestExtractStringFromFalkorDBFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "Three element array [id, len, str]",
			input:    []any{int64(1), int64(5), "hello"},
			expected: "hello",
		},
		{
			name:     "Three element array with bytes",
			input:    []any{int64(1), int64(5), []byte("world")},
			expected: "world",
		},
		{
			name:     "Two element array [id, str]",
			input:    []any{int64(1), "test"},
			expected: "test",
		},
		{
			name:     "Two element array with bytes",
			input:    []any{int64(1), []byte("test2")},
			expected: "test2",
		},
		{
			name:     "Direct string",
			input:    "direct",
			expected: "direct",
		},
		{
			name:     "Direct bytes",
			input:    []byte("bytes"),
			expected: "bytes",
		},
		{
			name:     "Empty array",
			input:    []any{},
			expected: "",
		},
		{
			name:     "Single element array",
			input:    []any{int64(1)},
			expected: "",
		},
		{
			name:     "Unsupported type",
			input:    123,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStringFromFalkorDBFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFalkorDBClose(t *testing.T) {
	t.Run("Close with valid client", func(t *testing.T) {
		fg, err := NewFalkorDBGraph("falkordb://localhost:6379/test")
		assert.NoError(t, err)
		assert.NotNil(t, fg)

		// Type assert to FalkorDBGraph
		graph := fg.(*FalkorDBGraph)
		err = graph.Close()
		// Close might fail if Redis is not running, but should not panic
		assert.NoError(t, err)
	})

	t.Run("Close with nil client", func(t *testing.T) {
		fg := &FalkorDBGraph{client: nil}
		err := fg.Close()
		assert.NoError(t, err)
	})
}

func TestInternalHelpers(t *testing.T) {
	t.Run("quoteString", func(t *testing.T) {
		assert.Equal(t, "\"test\"", quoteString("test"))
		assert.Equal(t, 123, quoteString(123))
		assert.Equal(t, true, quoteString(true))
	})

	t.Run("randomString", func(t *testing.T) {
		rs := randomString(10)
		assert.Len(t, rs, 10)

		// Different calls should produce different strings
		rs2 := randomString(10)
		// Note: Very small chance they could be equal, but extremely unlikely
		assert.NotEqual(t, rs, rs2)
	})

	t.Run("Node String", func(t *testing.T) {
		n := &Node{Alias: "a", Label: "Person", Properties: map[string]any{"name": "John"}}
		s := n.String()
		assert.Contains(t, s, "a:Person")
		assert.Contains(t, s, "name")
	})

	t.Run("Edge String", func(t *testing.T) {
		n1 := &Node{Alias: "a"}
		n2 := &Node{Alias: "b"}
		e := &Edge{Source: n1, Destination: n2, Relation: "KNOWS"}
		s := e.String()
		assert.Contains(t, s, "-[:KNOWS]->")
	})
}
