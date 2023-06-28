package parser

import (
	"fmt"
	"sort"
	"time"
	"unicode/utf8"

	sq "github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// Parser is a struct that holds the information needed to compose a sql query
// based on the user input parameters.
type Parser struct {
	Index  string
	Table  string
	Val    string
	Check  []CheckInfo
	Method Method
	Into   interface{}
}

// Method is a type that holds the method to be used in the sql query.
type Method struct {
	string
}

var (
	Update = Method{"update"}
	Insert = Method{"insert"}
	Delete = Method{"delete"}
	Query  = Method{"query"}
)

// CheckInfo is a helper struct to condense what is passed to the parser.
type CheckInfo struct {
	Fn      Prim
	Key     string
	Trimmed string
}

// Prim is a struct wrapping primitive types.
type Prim struct {
	I   int
	I64 int64
	S   string
	B   bool
	T   time.Time
}

type nameSortedCheck []CheckInfo

func (a nameSortedCheck) Len() int {
	return len(a)
}

func (a nameSortedCheck) Less(i, j int) bool {
	iRune, _ := utf8.DecodeRuneInString(a[i].Key)
	jRune, _ := utf8.DecodeRuneInString(a[j].Key)
	return iRune < jRune
}

func (a nameSortedCheck) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// ParseInput will take the input provided by the user via the table action methods
// and compose a sql query to be executed.
//
// As arguments are added they are sorted alphabetically. This is by no
// means a foolproof way of sorting the data but given the small subset of
// columns in our tables this will work to compose the arguments.
func ParseInput(p *Parser) (string, []interface{}, error) {
	valid := helpers.GetStructKeys(p.Into)

	if !validColumn(p.Index, valid) {
		return "", nil, fmt.Errorf("%s is not a valid column", p.Index)
	}

	for i := 0; i < len(p.Check); i++ {
		arg := p.Check[i]
		ok := indexNotArg(arg.Key, arg.Trimmed, arg.Fn, valid, p.Into)

		if !ok {
			p.Check = append(p.Check[:i], p.Check[i+1:]...)
			i--
		}
	}
	sort.Sort(nameSortedCheck(p.Check))

	base := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := p.builder(base)

	if err != nil {
		return sql, args, fmt.Errorf("sql generation failed %w", err)
	}

	return sql, args, err
}

func (p *Parser) builder(base sq.StatementBuilderType) (string, []interface{}, error) {
	switch p.Method {
	case Update:
		return p.update(base.Update(p.Table))
	case Delete:
		return p.delete(base.Delete(p.Table))
	default:
		return "", nil, nil
	}
}

func (p *Parser) delete(base sq.DeleteBuilder) (string, []interface{}, error) {
	for i := 0; i < len(p.Check); i++ {
		arg := p.Check[i]
		if arg.Fn.B {
			base.Where(sq.Eq{arg.Key: arg.Fn.B})
		}

		if !arg.Fn.T.IsZero() {
			base.Where(sq.Eq{arg.Key: arg.Fn.T.Format(time.RFC3339)})
		}

		if arg.Fn.S != "" {
			base.Where(sq.Eq{arg.Key: arg.Fn.S})
		}

		if arg.Fn.I != 0 {
			base.Where(sq.Eq{arg.Key: arg.Fn.I})
		}

		if arg.Fn.I64 != 0 {
			base.Where(sq.Eq{arg.Key: arg.Fn.I64})
		}
	}

	return base.ToSql()
}

func (p *Parser) update(base sq.UpdateBuilder) (string, []interface{}, error) {
	for i := 0; i < len(p.Check); i++ {
		arg := p.Check[i]
		if arg.Fn.B {
			base = base.Set(arg.Key, arg.Fn.B)
		}

		if !arg.Fn.T.IsZero() {
			base = base.Set(arg.Key, arg.Fn.T.Format(time.RFC3339))
		}

		if arg.Fn.S != "" {
			base = base.Set(arg.Key, arg.Fn.S)
		}

		if arg.Fn.I != 0 {
			base = base.Set(arg.Key, arg.Fn.I)
		}

		if arg.Fn.I64 != 0 {
			base = base.Set(arg.Key, arg.Fn.I64)
		}
	}

	base = base.Where(sq.Eq{p.Index: p.Val})

	return base.ToSql()
}

func indexNotArg(field, index string, opt Prim, opts []string, i interface{}) bool {
	if opt.S == "" && opt.I == 0 && opt.T.IsZero() && !opt.B {
		return false
	}

	fn, err := helpers.GetFieldName(field, i)
	if err != nil {
		return false
	}

	inputfn, err := helpers.GetFieldName(index, i)
	if err != nil {
		return false
	}

	// make sure the arg is not the index
	if fn == inputfn {
		return false
	}

	// make sure this is a valid option
	if !validColumn(field, opts) {
		return false
	}

	return true
}

func validColumn(s string, opts []string) bool {
	return helpers.Contains(opts, s)
}
