package splitter

import (
	"strings"
	"testing"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/stretchr/testify/assert"
)

func TestRecursiveCharacterTextSplitter(t *testing.T) {
	t.Run("Basic splitting", func(t *testing.T) {
		s := NewRecursiveCharacterTextSplitter(
			WithChunkSize(10),
			WithChunkOverlap(0),
		)
		text := "1234567890abcdefghij"
		chunks := s.SplitText(text)
		assert.Len(t, chunks, 2)
		assert.Equal(t, "1234567890", chunks[0])
		assert.Equal(t, "abcdefghij", chunks[1])
	})

	t.Run("Split with separators", func(t *testing.T) {
		s := NewRecursiveCharacterTextSplitter(
			WithChunkSize(10),
			WithChunkOverlap(0),
			WithSeparators([]string{"\n"}),
		)
		text := "part1\npart2\npart3"
		chunks := s.SplitText(text)
		assert.Len(t, chunks, 3)
		assert.Equal(t, "part1", chunks[0])
		assert.Equal(t, "part2", chunks[1])
		assert.Equal(t, "part3", chunks[2])
	})

	t.Run("Split documents", func(t *testing.T) {
		s := NewRecursiveCharacterTextSplitter(
			WithChunkSize(10),
			WithChunkOverlap(2),
		)
		doc := rag.Document{
			ID:       "doc1",
			Content:  "123456789012345",
			Metadata: map[string]any{"key": "val"},
		}
		chunks := s.SplitDocuments([]rag.Document{doc})

		assert.NotEmpty(t, chunks)
		for i, chunk := range chunks {
			assert.Equal(t, "doc1", chunk.Metadata["parent_id"])
			assert.Equal(t, i, chunk.Metadata["chunk_index"])
			assert.Equal(t, len(chunks), chunk.Metadata["chunk_total"])
		}
	})
}

func TestCharacterTextSplitter(t *testing.T) {
	s := NewCharacterTextSplitter(
		WithCharacterSeparator("|"),
		WithCharacterChunkSize(5),
		WithCharacterChunkOverlap(0),
	)
	text := "abc|def|ghi"
	chunks := s.SplitText(text)
	assert.Len(t, chunks, 3)
	assert.Equal(t, "abc", chunks[0])
	assert.Equal(t, "def", chunks[1])

	joined := s.JoinText(chunks)
	assert.Equal(t, "abc|def|ghi", joined)
}

func TestTokenTextSplitter(t *testing.T) {
	s := NewTokenTextSplitter(5, 0, nil)
	text := "one two three four five six seven eight"
	chunks := s.SplitText(text)
	assert.Len(t, chunks, 2)
	assert.Equal(t, "one two three four five", chunks[0])

	doc := rag.Document{ID: "tok1", Content: text}
	docChunks := s.SplitDocuments([]rag.Document{doc})
	assert.Len(t, docChunks, 2)
}

func TestRecursiveCharacterJoin(t *testing.T) {
	s := NewRecursiveCharacterTextSplitter(WithChunkOverlap(0))
	joined := s.JoinText([]string{"a", "b"})
	assert.Equal(t, "a b", joined)
}

// SimpleTextSplitter tests
func TestSimpleTextSplitter(t *testing.T) {
	t.Run("NewSimpleTextSplitter", func(t *testing.T) {
		s := NewSimpleTextSplitter(100, 10)
		assert.NotNil(t, s)
		simple := s.(*SimpleTextSplitter)
		assert.Equal(t, 100, simple.ChunkSize)
		assert.Equal(t, 10, simple.ChunkOverlap)
		assert.Equal(t, "\n\n", simple.Separator)
	})

	t.Run("SplitText short text", func(t *testing.T) {
		s := NewSimpleTextSplitter(100, 10)
		text := "Short text"
		chunks := s.SplitText(text)
		assert.Len(t, chunks, 1)
		assert.Equal(t, "Short text", chunks[0])
	})

	t.Run("SplitText exact size", func(t *testing.T) {
		s := NewSimpleTextSplitter(10, 0)
		text := "1234567890" // exactly 10 chars
		chunks := s.SplitText(text)
		assert.Len(t, chunks, 1)
		assert.Equal(t, "1234567890", chunks[0])
	})

	t.Run("SplitText multiple chunks", func(t *testing.T) {
		s := NewSimpleTextSplitter(10, 0)
		text := "1234567890abcdefghijklmnop"
		chunks := s.SplitText(text)
		assert.Len(t, chunks, 3)
		assert.Equal(t, "1234567890", chunks[0])
		assert.Equal(t, "abcdefghij", chunks[1])
		assert.Equal(t, "klmnop", chunks[2])
	})

	t.Run("SplitText with separator", func(t *testing.T) {
		s := NewSimpleTextSplitter(20, 0)
		// Default separator is "\n\n"
		text := "First paragraph\n\nSecond paragraph here"
		chunks := s.SplitText(text)
		// Text is 43 chars, chunk size is 20
		// It will split into multiple chunks
		assert.Greater(t, len(chunks), 1)
		assert.Contains(t, chunks[0], "First paragraph")
	})

	t.Run("SplitText with overlap", func(t *testing.T) {
		s := NewSimpleTextSplitter(20, 5)
		text := "12345678901234567890abcdefghijklmnopqrstuvwxyz"
		chunks := s.SplitText(text)
		assert.Greater(t, len(chunks), 1)
		// Check that there is overlap between consecutive chunks
		if len(chunks) > 1 {
			// The last 5 chars of chunk 0 should appear in chunk 1
			endOfFirst := chunks[0][len(chunks[0])-5:]
			assert.Contains(t, chunks[1], endOfFirst)
		}
	})

	t.Run("SplitText empty string", func(t *testing.T) {
		s := NewSimpleTextSplitter(100, 10)
		chunks := s.SplitText("")
		assert.Len(t, chunks, 1)
		assert.Equal(t, "", chunks[0])
	})

	t.Run("SplitText with very small chunk size", func(t *testing.T) {
		s := NewSimpleTextSplitter(3, 0)
		text := "abcdefgh"
		chunks := s.SplitText(text)
		assert.Len(t, chunks, 3)
		assert.Equal(t, "abc", chunks[0])
		assert.Equal(t, "def", chunks[1])
		assert.Equal(t, "gh", chunks[2])
	})

	t.Run("SplitText overlap prevents getting stuck", func(t *testing.T) {
		s := NewSimpleTextSplitter(10, 8)
		text := "12345678901234567890"
		chunks := s.SplitText(text)
		assert.Greater(t, len(chunks), 1)
	})

	t.Run("JoinText empty chunks", func(t *testing.T) {
		s := NewSimpleTextSplitter(100, 10)
		joined := s.JoinText([]string{})
		assert.Equal(t, "", joined)
	})

	t.Run("JoinText single chunk", func(t *testing.T) {
		s := NewSimpleTextSplitter(100, 10)
		joined := s.JoinText([]string{"single"})
		assert.Equal(t, "single", joined)
	})

	t.Run("JoinText multiple chunks", func(t *testing.T) {
		s := NewSimpleTextSplitter(100, 10)
		joined := s.JoinText([]string{"first", "second", "third"})
		assert.Equal(t, "first second third", joined)
	})

	t.Run("SplitDocuments single document", func(t *testing.T) {
		s := NewSimpleTextSplitter(20, 5)
		doc := rag.Document{
			ID:       "doc1",
			Content:  "This is a test document that should be split into multiple chunks for testing",
			Metadata: map[string]any{"source": "test"},
		}
		chunks := s.SplitDocuments([]rag.Document{doc})
		assert.Greater(t, len(chunks), 1)

		// Verify metadata
		for i, chunk := range chunks {
			assert.Equal(t, "doc1", chunk.ID)
			assert.Equal(t, "test", chunk.Metadata["source"])
			assert.Equal(t, i, chunk.Metadata["chunk_index"])
			assert.Equal(t, len(chunks), chunk.Metadata["total_chunks"])
		}
	})

	t.Run("SplitDocuments multiple documents", func(t *testing.T) {
		s := NewSimpleTextSplitter(15, 0)
		docs := []rag.Document{
			{ID: "doc1", Content: "First document content"},
			{ID: "doc2", Content: "Second document here"},
		}
		chunks := s.SplitDocuments(docs)
		assert.Greater(t, len(chunks), 2)

		// First doc chunks should have doc1 as parent_id
		doc1Chunks := 0
		doc2Chunks := 0
		for _, chunk := range chunks {
			if chunk.ID == "doc1" {
				doc1Chunks++
			} else if chunk.ID == "doc2" {
				doc2Chunks++
			}
		}
		assert.Greater(t, doc1Chunks, 0)
		assert.Greater(t, doc2Chunks, 0)
	})

	t.Run("SplitText with custom separator", func(t *testing.T) {
		s := &SimpleTextSplitter{
			ChunkSize:    15,
			ChunkOverlap: 0,
			Separator:    "||",
		}
		text := "Part1||Part2||Part3"
		chunks := s.SplitText(text)
		// Text is 21 chars, chunk size is 15
		// Will split: first chunk to "Part1||Part2||" then second
		assert.GreaterOrEqual(t, len(chunks), 1)
		assert.Contains(t, chunks[0], "Part1")
	})

	t.Run("SplitText trims whitespace", func(t *testing.T) {
		s := NewSimpleTextSplitter(10, 0)
		text := "1234567890   abcdefghij   "
		chunks := s.SplitText(text)
		assert.Greater(t, len(chunks), 1)
		// Chunks should be trimmed
		for _, chunk := range chunks {
			assert.Equal(t, strings.TrimSpace(chunk), chunk)
		}
	})
}
