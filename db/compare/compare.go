package compare

import (
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Compare struct {
	val string
}

var (
	Equal              = Compare{val: "="}
	GreaterThan        = Compare{val: ">"}
	GreaterThanOrEqual = Compare{val: ">="}
	LessThan           = Compare{val: "<"}
	LessThanOrEqual    = Compare{val: "<="}
)

// Comparison is a helper function to build a query with a comparison operator.
func Comparison(s sq.SelectBuilder, column, val string, c Compare) sq.SelectBuilder {
	switch c {
	case Equal:
		return s.Where(sq.Eq{column: val})
	case GreaterThan:
		return s.Where(sq.Gt{column: val})
	case GreaterThanOrEqual:
		return s.Where(sq.GtOrEq{column: val})
	case LessThan:
		return s.Where(sq.Lt{column: val})
	case LessThanOrEqual:
		return s.Where(sq.LtOrEq{column: val})
	default:
		return s.Where(sq.Eq{column: val})
	}
}

// EmptyOrTime checks if the time.Time object is zero and returns nil if it is.
func EmptyOrTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

// ZeroNil checks if the given int is zero and returns nil if it is.
func ZeroNil(i *int) any {
	if *i == 0 {
		return nil
	}
	return i
}
