package retriever

import (
	"context"
	"math"
	"sort"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/tokenizer"
)

// BM25Retriever implements BM25 (Best Matching 25) sparse retrieval
// BM25 is a ranking function based on probabilistic information retrieval
type BM25Retriever struct {
	documents []rag.Document
	tokenizer tokenizer.Tokenizer
	config    BM25Config
	index     *bm25Index
}

// BM25Config contains configuration parameters for BM25
type BM25Config struct {
	// k1 controls term frequency saturation
	// Higher k1 gives more weight to term frequency
	// Typical values: 1.2 - 2.0
	K1 float64

	// b controls document length normalization
	// 0 = no length normalization, 1 = full normalization
	// Typical value: 0.75
	B float64

	// k is the number of documents to retrieve
	K int

	// ScoreThreshold filters results by minimum score
	ScoreThreshold float64
}

// DefaultBM25Config returns default BM25 configuration
func DefaultBM25Config() BM25Config {
	return BM25Config{
		K1:             1.5,
		B:              0.75,
		K:              4,
		ScoreThreshold: 0.0,
	}
}

// bm25Index stores the inverted index and document statistics
type bm25Index struct {
	// Inverted index: term -> list of (docID, termFrequency)
	termFreqs map[string][]docTermFreq

	// Document frequencies: term -> number of documents containing term
	docFreqs map[string]int

	// Document lengths
	docLengths []int

	// Average document length
	avgDocLength float64

	// Number of documents
	numDocs int
}

// docTermFreq stores term frequency for a document
type docTermFreq struct {
	docID    int
	termFreq int
}

// NewBM25Retriever creates a new BM25 retriever from documents
func NewBM25Retriever(documents []rag.Document, config BM25Config) (*BM25Retriever, error) {
	if config.K1 == 0 {
		config.K1 = 1.5
	}
	if config.B == 0 {
		config.B = 0.75
	}
	if config.K == 0 {
		config.K = 4
	}

	// Use default tokenizer if none provided
	tok := tokenizer.DefaultRegexTokenizer()

	retriever := &BM25Retriever{
		documents: documents,
		tokenizer: tok,
		config:    config,
	}

	// Build the index
	retriever.buildIndex()

	return retriever, nil
}

// NewBM25RetrieverWithTokenizer creates a BM25 retriever with custom tokenizer
func NewBM25RetrieverWithTokenizer(documents []rag.Document, config BM25Config, tok tokenizer.Tokenizer) (*BM25Retriever, error) {
	if config.K1 == 0 {
		config.K1 = 1.5
	}
	if config.B == 0 {
		config.B = 0.75
	}
	if config.K == 0 {
		config.K = 4
	}

	retriever := &BM25Retriever{
		documents: documents,
		tokenizer: tok,
		config:    config,
	}

	retriever.buildIndex()

	return retriever, nil
}

// buildIndex builds the BM25 inverted index
func (r *BM25Retriever) buildIndex() {
	index := &bm25Index{
		termFreqs:  make(map[string][]docTermFreq),
		docFreqs:   make(map[string]int),
		docLengths: make([]int, len(r.documents)),
		numDocs:    len(r.documents),
	}

	totalLength := 0

	// Build term frequency index
	for docID, doc := range r.documents {
		tokens := r.tokenizer.Tokenize(doc.Content)
		docLen := len(tokens)
		index.docLengths[docID] = docLen
		totalLength += docLen

		// Count term frequencies in this document
		termCounts := make(map[string]int)
		for _, token := range tokens {
			termCounts[token]++
		}

		// Update inverted index
		for term, freq := range termCounts {
			index.termFreqs[term] = append(index.termFreqs[term], docTermFreq{
				docID:    docID,
				termFreq: freq,
			})
			index.docFreqs[term]++
		}
	}

	// Calculate average document length
	if index.numDocs > 0 {
		index.avgDocLength = float64(totalLength) / float64(index.numDocs)
	}

	r.index = index
}

// Retrieve retrieves documents based on a query
func (r *BM25Retriever) Retrieve(ctx context.Context, query string) ([]rag.Document, error) {
	return r.RetrieveWithK(ctx, query, r.config.K)
}

// RetrieveWithK retrieves exactly k documents
func (r *BM25Retriever) RetrieveWithK(ctx context.Context, query string, k int) ([]rag.Document, error) {
	bm25Config := r.config
	bm25Config.K = k
	retrievalConfig := &rag.RetrievalConfig{
		K:              bm25Config.K,
		ScoreThreshold: bm25Config.ScoreThreshold,
	}
	results, err := r.RetrieveWithConfig(ctx, query, retrievalConfig)
	if err != nil {
		return nil, err
	}

	docs := make([]rag.Document, len(results))
	for i, result := range results {
		docs[i] = result.Document
	}

	return docs, nil
}

// RetrieveWithConfig retrieves documents with custom configuration
func (r *BM25Retriever) RetrieveWithConfig(ctx context.Context, query string, config *rag.RetrievalConfig) ([]rag.DocumentSearchResult, error) {
	if config == nil {
		cfg := r.config
		config = &rag.RetrievalConfig{
			K:              cfg.K,
			ScoreThreshold: cfg.ScoreThreshold,
		}
	}

	// Tokenize query
	queryTokens := r.tokenizer.Tokenize(query)
	if len(queryTokens) == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	// Calculate BM25 scores for each document
	scores := make([]float64, len(r.documents))
	hasMatch := make([]bool, len(r.documents))

	// For each query term, calculate its contribution to document scores
	for _, term := range queryTokens {
		// Skip terms not in index
		docFreq, exists := r.index.docFreqs[term]
		if !exists || docFreq == 0 {
			continue
		}

		// Calculate IDF (Inverse Document Frequency)
		// IDF = log((N - df + 0.5) / (df + 0.5) + 1)
		// Using a variant that prevents negative values
		idf := math.Log((float64(r.index.numDocs)-float64(docFreq)+0.5)/(float64(docFreq)+0.5) + 1.0)

		// Get documents containing this term
		termDocs := r.index.termFreqs[term]

		for _, td := range termDocs {
			docID := td.docID
			termFreq := td.termFreq
			docLength := r.index.docLengths[docID]

			// Calculate BM25 component for this term
			// TF component: (freq * (k1 + 1)) / (freq + k1 * (1 - b + b * (docLen / avgDocLen)))
			numerator := float64(termFreq) * (r.config.K1 + 1.0)
			denominator := float64(termFreq) + r.config.K1*(1.0-r.config.B+r.config.B*(float64(docLength)/r.index.avgDocLength))

			tfComponent := numerator / denominator

			// Add to document score
			scores[docID] += idf * tfComponent
			hasMatch[docID] = true
		}
	}

	// Create results from scored documents
	results := make([]rag.DocumentSearchResult, 0)
	for docID, score := range scores {
		if !hasMatch[docID] {
			continue
		}

		// Filter by score threshold
		if config.ScoreThreshold > 0 && score < config.ScoreThreshold {
			continue
		}

		results = append(results, rag.DocumentSearchResult{
			Document: r.documents[docID],
			Score:    score,
			Metadata: map[string]any{
				"retriever_type": "bm25",
			},
		})
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit to K results
	if len(results) > config.K {
		results = results[:config.K]
	}

	return results, nil
}

// AddDocuments adds new documents to the index
func (r *BM25Retriever) AddDocuments(docs []rag.Document) {
	// Append to documents
	r.documents = append(r.documents, docs...)

	// Rebuild index
	r.buildIndex()
}

// DeleteDocument removes a document by ID
func (r *BM25Retriever) DeleteDocument(docID string) {
	// Find and remove document
	for i, doc := range r.documents {
		if doc.ID == docID {
			r.documents = append(r.documents[:i], r.documents[i+1:]...)
			break
		}
	}

	// Rebuild index
	r.buildIndex()
}

// UpdateDocument updates a document in the index
func (r *BM25Retriever) UpdateDocument(doc rag.Document) {
	// Find and update document
	for i, d := range r.documents {
		if d.ID == doc.ID {
			r.documents[i] = doc
			break
		}
	}

	// Rebuild index
	r.buildIndex()
}

// GetDocumentCount returns the number of indexed documents
func (r *BM25Retriever) GetDocumentCount() int {
	return len(r.documents)
}

// GetStats returns statistics about the index
func (r *BM25Retriever) GetStats() map[string]any {
	return map[string]any{
		"num_documents":    len(r.documents),
		"num_unique_terms": len(r.index.docFreqs),
		"avg_doc_length":   r.index.avgDocLength,
		"k1":               r.config.K1,
		"b":                r.config.B,
	}
}

// SetK1 updates the k1 parameter and rebuilds the index
func (r *BM25Retriever) SetK1(k1 float64) {
	r.config.K1 = k1
	// No need to rebuild index, just parameter change
}

// SetB updates the b parameter and rebuilds the index
func (r *BM25Retriever) SetB(b float64) {
	r.config.B = b
	// No need to rebuild index, just parameter change
}

// GetK1 returns the current k1 parameter
func (r *BM25Retriever) GetK1() float64 {
	return r.config.K1
}

// GetB returns the current b parameter
func (r *BM25Retriever) GetB() float64 {
	return r.config.B
}
