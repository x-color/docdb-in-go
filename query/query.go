package query

import (
	"fmt"
	"strconv"
)

type kind string

const (
	kindKey   kind = "key"
	kindValue kind = "value"
	kindOp    kind = "op"
	kindEOF   kind = "EOF"
)

type token struct {
	kind  kind
	value string
}

func newToken(kind kind, value string) token {
	return token{
		kind:  kind,
		value: value,
	}
}

type lexer struct {
	input string
	index int
	kind  kind
}

func (l lexer) process() ([]token, error) {
	tokens := make([]token, 0)
	for len(tokens) < 10 {
		token, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
		if token.kind == kindEOF {
			break
		}
	}
	return tokens, nil
}

func (l *lexer) nextToken() (token, error) {
	for {
		switch l.peekChar() {
		case '"':
			l.readChar()
			str, err := l.readString()
			if err != nil {
				return token{}, err
			}
			return newToken(l.kind, str), nil
		case ':':
			l.readChar()
			l.switchKind()
			return l.operator()
		case '.':
			l.readChar()
		case ' ':
			l.skipSpace()
			l.switchKind()
		case 0:
			return newToken(kindEOF, "EOF"), nil
		default:
			return newToken(l.kind, l.readWord()), nil
		}
	}
}

func (l *lexer) readWord() string {
	l.index++
	i := l.index
	for {
		ch := l.readChar()
		if l.isSpecialChar(ch) {
			l.index--
			break
		}
	}
	return l.input[i : l.index+1]
}

func (l lexer) isSpecialChar(ch byte) bool {
	return ch == '"' || ch == ':' || ch == '.' || ch == ' ' || ch == 0
}

func (l *lexer) skipSpace() {
	for l.peekChar() == ' ' {
		l.readChar()
	}
}

func (l *lexer) operator() (token, error) {
	switch l.peekChar() {
	case '>':
		l.readChar()
		return newToken(kindOp, ">"), nil
	case '<':
		l.readChar()
		return newToken(kindOp, "<"), nil
	case 0:
		return token{}, fmt.Errorf("unexpected character at %d", l.index)
	default:
		return newToken(kindOp, "="), nil
	}
}

func (l *lexer) readString() (string, error) {
	l.index++
	i := l.index
	for {
		ch := l.readChar()
		if ch == '"' {
			break
		}
		if ch == 0 {
			return "", fmt.Errorf("unexpected character at %d", l.index)
		}
	}
	return l.input[i:l.index], nil
}

func (l *lexer) readChar() byte {
	l.index++
	if l.index >= len(l.input) {
		return 0
	}
	return l.input[l.index]
}

func (l lexer) peekChar() byte {
	l.index++
	if l.index >= len(l.input) {
		return 0
	}
	return l.input[l.index]
}

func (l *lexer) switchKind() {
	switch l.kind {
	case kindKey:
		l.kind = kindValue
	case kindValue:
		l.kind = kindKey
	}
}

func newLexer(input string) lexer {
	l := lexer{
		input: input,
		kind:  kindKey,
		index: -1,
	}
	l.skipSpace()
	return l
}

type operation string

const (
	opeEq operation = "="
	opeLt operation = "<"
	opeGt operation = ">"
)

type query struct {
	Keys  []string
	Value string
	Op    operation
}

type Queries []query

func ParseQuery(rq string) (Queries, error) {
	if rq == "" {
		return nil, nil
	}

	l := newLexer(rq)
	tokens, err := l.process()
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	queries := make(Queries, 0)
	q := query{}
	for _, token := range tokens {
		switch token.kind {
		case kindKey:
			q.Keys = append(q.Keys, token.value)
		case kindOp:
			q.Op = operation(token.value)
		case kindValue:
			q.Value = token.value
			queries = append(queries, q)
			q = query{}
		}
	}

	if err := queries.validate(); err != nil {
		return nil, err
	}

	return queries, nil
}

func (qs Queries) validate() error {
	if len(qs) == 0 {
		return fmt.Errorf("invalid query")
	}
	for _, q := range qs {
		if len(q.Keys) == 0 || len(string(q.Op)) == 0 || len(q.Value) == 0 {
			return fmt.Errorf("invalid query")
		}
	}
	return nil
}

func (qs Queries) Match(doc map[string]any) bool {
	for _, q := range qs {
		v := qs.get(doc, q.Keys)
		if v == nil {
			return false
		}

		if q.Op == opeEq {
			if q.Value == fmt.Sprintf("%v", v) {
				continue
			}
		}

		r, err := strconv.ParseFloat(q.Value, 64)
		if err != nil {
			return false
		}
		var l float64
		switch t := v.(type) {
		case float64:
			l = t
		case float32:
			l = float64(t)
		case uint:
			l = float64(t)
		case uint8:
			l = float64(t)
		case uint16:
			l = float64(t)
		case uint32:
			l = float64(t)
		case uint64:
			l = float64(t)
		case int:
			l = float64(t)
		case int8:
			l = float64(t)
		case int16:
			l = float64(t)
		case int32:
			l = float64(t)
		case int64:
			l = float64(t)
		case string:
			l, err = strconv.ParseFloat(t, 64)
			if err != nil {
				return false
			}
		default:
			return false
		}

		if q.Op == ">" {
			if l <= r {
				return false
			}
			continue
		}

		if l >= r {
			return false
		}
	}
	return true
}

func (qs Queries) get(doc map[string]any, path []string) any {
	for i, k := range path {
		v, ok := doc[k]
		if !ok {
			return nil
		}
		if i == len(path)-1 {
			break
		}
		doc, ok = v.(map[string]any)
		if !ok {
			return nil
		}
	}

	v := doc[path[len(path)-1]]
	if _, ok := v.(map[string]any); ok {
		return nil
	}

	return v
}
