package search

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
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

// Search performs a simple query across indexed documents.
func (i *Index) Search(query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	req := bleve.NewSearchRequest(bleve.NewQueryStringQuery(query))
	req.Size = limit
	result, err := i.index.Search(req)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		ids = append(ids, hit.ID)
	}
	return ids, nil
}

// Close closes the index.
func (i *Index) Close() error {
	return i.index.Close()
}

// Path returns the default index path under site dir.
func Path(siteDir string) string {
	return filepath.Join(siteDir, "index", "search.bleve")
}