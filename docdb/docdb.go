package docdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

var (
	ErrUnknown  = errors.New("unknown error")
	ErrFatal    = errors.New("fatal error")
	ErrNotFound = errors.New("not found error")
)

type DocDBError struct {
	typ error
	msg string
}

func (e DocDBError) Error() string {
	return e.msg
}

func (e DocDBError) Is(target error) bool {
	return e.typ == target
}

func wrapError(typ error, msg string) error {
	return DocDBError{
		typ: typ,
		msg: msg,
	}
}

type DocDB struct {
	db *cache.Cache
}

func (d DocDB) Add(doc map[string]any) (string, error) {
	id := uuid.New().String()
	b, err := json.Marshal(doc)
	if err != nil {
		return "", wrapError(ErrFatal, fmt.Sprintf("failed to convert document to byte data: %s", err))
	}

	d.db.Set(id, b, 0)
	return id, nil
}

func (d DocDB) Get(id string) (map[string]any, error) {
	item, ok := d.db.Get(id)
	if !ok {
		return nil, wrapError(ErrNotFound, fmt.Sprintf("not found document by %s", id))
	}
	b, ok := item.([]byte)
	if !ok {
		return nil, wrapError(ErrFatal, fmt.Sprintf("unexpected data in %s", id))
	}
	doc := make(map[string]any)
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, wrapError(ErrFatal, fmt.Sprintf("failed to convert data to document: %s", err))
	}

	return doc, nil
}

func (d DocDB) GetAll() (map[string]map[string]any, error) {
	docs := make(map[string]map[string]any)
	items := d.db.Items()
	for id, item := range items {
		b, ok := item.Object.([]byte)
		if !ok {
			return nil, wrapError(ErrFatal, fmt.Sprintf("unexpected data in %s", id))
		}
		doc := make(map[string]any)
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, wrapError(ErrFatal, fmt.Sprintf("failed to convert data to document: %s", err))
		}
		docs[id] = doc
	}
	return docs, nil
}

func NewDocDB() *DocDB {
	return &DocDB{
		db: cache.New(30*time.Minute, 10*time.Minute),
	}
}
