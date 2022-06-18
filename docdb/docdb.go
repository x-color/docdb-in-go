package docdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/x-color/docdb-in-go/query"
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
	db      *cache.Cache
	indexDb *cache.Cache
}

func (d DocDB) Add(doc map[string]any) (string, error) {
	id := uuid.New().String()
	b, err := json.Marshal(doc)
	if err != nil {
		return "", wrapError(ErrFatal, fmt.Sprintf("failed to convert document to byte data: %s", err))
	}

	d.db.Set(id, b, 0)
	d.index(id, doc)

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

func (d DocDB) Search(q query.Queries) ([]map[string]any, error) {
	match := make([]map[string]any, 0)
	for id, item := range d.db.Items() {
		b, ok := item.Object.([]byte)
		if !ok {
			return nil, wrapError(ErrFatal, fmt.Sprintf("unexpected data in %s", id))
		}
		doc := make(map[string]any)
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, wrapError(ErrFatal, fmt.Sprintf("failed to convert data to document: %s", err))
		}
		if q.Match(doc) {
			match = append(match, map[string]any{
				"id":       id,
				"document": doc,
			})
		}
	}
	return match, nil
}

func (d DocDB) index(id string, doc map[string]any) {
	pvs := getPathValues(doc, "")
	for _, pv := range pvs {
		v, ok := d.indexDb.Get(pv)
		if !ok {
			continue
		}
		ids, ok := v.(string)
		if !ok {
			continue
		}

		if !strings.Contains(id, ids) {
			ids = fmt.Sprintf("%s,%s", ids, id)
			d.indexDb.Set(pv, []byte(ids), 0)
		}
	}
}

func getPathValues(obj map[string]any, prefix string) []string {
	var pvs []string
	for k, v := range obj {
		if prefix != "" {
			k = prefix + "." + k
		}
		switch t := v.(type) {
		case map[string]any:
			pvs = append(pvs, getPathValues(t, k)...)
			continue
		case []any:
			continue
		}

		pvs = append(pvs, fmt.Sprintf("%s=%v", k, v))
	}

	return pvs
}

func NewDocDB() *DocDB {
	return &DocDB{
		db:      cache.New(30*time.Minute, 10*time.Minute),
		indexDb: cache.New(30*time.Minute, 10*time.Minute),
	}
}
