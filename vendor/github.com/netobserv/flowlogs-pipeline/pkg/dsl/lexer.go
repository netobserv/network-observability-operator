package dsl

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/netobserv/flowlogs-pipeline/pkg/utils/filters"
)

const (
	operatorOr  = "or"
	operatorAnd = "and"
)

var syntaxTokens = map[string]int{
	"=":         EQ,
	"!=":        NEQ,
	"=~":        REG,
	"!~":        NREG,
	">":         GT,
	"<":         LT,
	">=":        GE,
	"<=":        LE,
	operatorOr:  OR,
	operatorAnd: AND,
	"with":      WITH,
	"without":   WITHOUT,
	"(":         OPEN_PARENTHESIS,
	")":         CLOSE_PARENTHESIS,
}

type Expression interface {
	toTree() (*tree, error)
}

type ParenthesisExpr struct {
	inner Expression
}

func (e ParenthesisExpr) toTree() (*tree, error) {
	return e.inner.toTree()
}

type kvPair struct {
	key   string
	value string
}

type kvPairInt struct {
	key   string
	value int
}

type EqExpr kvPair

func (e EqExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.Equal(e.key, e.value, false)}, nil
}

type NEqExpr kvPair

func (e NEqExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.NotEqual(e.key, e.value, false)}, nil
}

type EqNumExpr kvPairInt

func (e EqNumExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.NumEquals(e.key, e.value)}, nil
}

type NEqNumExpr kvPairInt

func (e NEqNumExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.NumNotEquals(e.key, e.value)}, nil
}

type LessThanExpr kvPairInt

func (e LessThanExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.LessThan(e.key, e.value)}, nil
}

type GreaterThanExpr kvPairInt

func (e GreaterThanExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.GreaterThan(e.key, e.value)}, nil
}

type LessOrEqualThanExpr kvPairInt

func (e LessOrEqualThanExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.LessOrEqualThan(e.key, e.value)}, nil
}

type GreaterOrEqualThanExpr kvPairInt

func (e GreaterOrEqualThanExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.GreaterOrEqualThan(e.key, e.value)}, nil
}

type RegExpr kvPair

func (e RegExpr) toTree() (*tree, error) {
	r, err := regexp.Compile(e.value)
	if err != nil {
		return nil, fmt.Errorf("invalid regex filter: cannot compile regex [%w]", err)
	}
	return &tree{predicate: filters.Regex(e.key, r)}, nil
}

type NRegExpr kvPair

func (e NRegExpr) toTree() (*tree, error) {
	r, err := regexp.Compile(e.value)
	if err != nil {
		return nil, fmt.Errorf("invalid regex filter: cannot compile regex [%w]", err)
	}
	return &tree{predicate: filters.NotRegex(e.key, r)}, nil
}

type WithExpr struct {
	key string
}

func (e WithExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.Presence(e.key)}, nil
}

type WithoutExpr struct {
	key string
}

func (e WithoutExpr) toTree() (*tree, error) {
	return &tree{predicate: filters.Absence(e.key)}, nil
}

type LogicalExpr struct {
	left     Expression
	operator string
	right    Expression
}

func (le LogicalExpr) toTree() (*tree, error) {
	left, err := le.left.toTree()
	if err != nil {
		return nil, err
	}
	right, err := le.right.toTree()
	if err != nil {
		return nil, err
	}
	return &tree{
		logicalOp: le.operator,
		children:  []*tree{left, right},
	}, nil
}

type Lexer struct {
	scanner.Scanner
	errs   []error
	result Expression
}

func (l *Lexer) Lex(lval *yySymType) int {
	token := l.Scan()
	if token == scanner.EOF {
		return 0
	}
	tokenText := l.TokenText()
	lval.value = tokenText

	switch token {
	case scanner.Int:
		// Reading arbitrary number
		res, err := strconv.ParseInt(tokenText, 10, 64)
		if err != nil {
			l.Error(err.Error())
			return 0
		}
		lval.intValue = int(res)
		return NUMBER
	case scanner.Float:
		l.Error("Float values are currently unsupported")
		return 0
	case scanner.String, scanner.RawString:
		// Reading arbitrary double-quotes delimited string
		var err error
		lval.value, err = strconv.Unquote(tokenText)
		if err != nil {
			l.Error(err.Error())
			return 0
		}
		return STRING
	}

	// Check if this is a syntaxToken

	// Some characters are read as a token, such as "=", regardless of what follows
	// To read "=~" as a token, we need to Peek next rune manually
	tokenNext := tokenText + string(l.Peek())
	if tok, ok := syntaxTokens[tokenNext]; ok {
		l.Next()
		return tok
	}

	if tok, ok := syntaxTokens[strings.ToLower(tokenText)]; ok {
		return tok
	}

	// When none of the above returned, this must be a NetFlow field name
	return NF_FIELD
}

func (l *Lexer) Error(msg string) {
	l.errs = append(l.errs, fmt.Errorf("%s: %d:%d", msg, l.Line, l.Column))
}

func Parse(s string) (filters.Predicate, error) {
	l := new(Lexer)
	l.Init(strings.NewReader(s))
	yyErrorVerbose = true
	yyParse(l)
	if len(l.errs) > 0 {
		return nil, errors.Join(l.errs...)
	}
	t, err := l.result.toTree()
	if err != nil {
		return nil, err
	}
	return t.apply, nil
}
