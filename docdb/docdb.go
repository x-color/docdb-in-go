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

func (d DocDB) Search(qs query.Queries) ([]map[string]any, error) {
	matchId := make(map[string]int)
	for _, q := range qs {
		if q.Op == query.OpeEq {
			ids, err := d.lookup(fmt.Sprintf("%s=%s", strings.Join(q.Keys, "."), q.Value))
			if err != nil {
				return nil, wrapError(ErrFatal, fmt.Sprintf("failed to get data from index: %v", q))
			}
			for _, id := range ids {
				matchId[id]++
			}
		} else {
			ids, err := d.lookup(strings.Join(q.Keys, "."))
			if err != nil {
				return nil, wrapError(ErrFatal, fmt.Sprintf("failed to get data from index: %v", q))
			}
			for _, id := range ids {
				matchId[id]++
			}
		}
	}

	match := make([]map[string]any, 0)
	for id, count := range matchId {
		if count != len(qs) {
			continue
		}
		doc, err := d.Get(id)
		if err != nil {
			return nil, wrapError(ErrFatal, fmt.Sprintf("failed to get doc from main: %s", id))
		}
		if qs.Match(doc) {
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
	d.setIndex(id, pvs)
	ps := getPath(doc, "")
	d.setIndex(id, ps)
}

func (d DocDB) setIndex(id string, keys []string) {
	for _, key := range keys {
		v, ok := d.indexDb.Get(key)
		if !ok {
			d.indexDb.Set(key, id, 0)
			continue
		}
		ids, ok := v.(string)
		if !ok {
			continue
		}

		if !strings.Contains(id, ids) {
			ids = fmt.Sprintf("%s,%s", ids, id)
			d.indexDb.Set(key, ids, 0)
		}
	}
}

func (d DocDB) lookup(pv string) ([]string, error) {
	b, ok := d.indexDb.Get(pv)
	if !ok {
		return nil, nil
	}
	s, ok := b.(string)
	if !ok {
		return nil, wrapError(ErrFatal, fmt.Sprintf("failed to convert data in indexDB to string: %v", pv))
	}
	return strings.Split(s, ","), nil
}

func getPath(obj map[string]any, prefix string) []string {
	var path []string
	for k, v := range obj {
		if prefix != "" {
			k = prefix + "." + k
		}
		switch t := v.(type) {
		case map[string]any:
			path = append(path, getPath(t, k)...)
			continue
		case []any:
			continue
		}

		path = append(path, k)
	}

	return path
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
