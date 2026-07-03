package search

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
)

// Index wraps a Bleve full-text index for code and entity search.
type Index struct {
	index bleve.Index
}

// Open opens or creates a search index at path.
func Open(path string) (*Index, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(path, mapping)
	}
	if err != nil {
		return nil, err
	}
	return &Index{index: index}, nil
}

// Document is a searchable entity.
type Document struct {
	ID      string
	Type    string
	Project string
	Title   string
	Body    string
}

// IndexDocument indexes or replaces a document.
func (i *Index) IndexDocument(doc Document) error {
	return i.index.Index(doc.ID, doc)
}

// SearchDocuments performs a query and returns matching documents.
func (i *Index) SearchDocuments(query string, limit int) ([]Document, error) {
	if limit <= 0 {
		limit = 20
	}
	req := bleve.NewSearchRequest(bleve.NewQueryStringQuery(query))
	req.Size = limit
	req.Fields = []string{"*"}
	result, err := i.index.Search(req)
	if err != nil {
		return nil, err
	}
	docs := make([]Document, 0, len(result.Hits))
	for _, hit := range result.Hits {
		docs = append(docs, documentFromHit(hit))
	}
	return docs, nil
}

func documentFromHit(hit *search.DocumentMatch) Document {
	doc := Document{ID: hit.ID}
	if v, ok := hit.Fields["Type"].(string); ok {
		doc.Type = v
	}
	if v, ok := hit.Fields["Project"].(string); ok {
		doc.Project = v
	}
	if v, ok := hit.Fields["Title"].(string); ok {
		doc.Title = v
	}
	if v, ok := hit.Fields["Body"].(string); ok {
		doc.Body = v
	}
	return doc
}

// Close closes the index.
func (i *Index) Close() error {
	return i.index.Close()
}

// Path returns the default index path under site dir.
func Path(siteDir string) string {
	return filepath.Join(siteDir, "index", "search.bleve")
}