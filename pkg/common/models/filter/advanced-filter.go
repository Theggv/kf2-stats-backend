package filter

import (
	"fmt"

	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

type AdvancedFilterOp = int

const (
	Eq AdvancedFilterOp = iota + 1
	NotEq
	Gt
	Gte
	Lt
	Lte
	In
	NotIn
	Between
)

type AdvancedFilter struct {
	Op   AdvancedFilterOp `json:"operator"`
	Args []float64        `json:"args"`
}

func (c AdvancedFilter) validate() bool {
	switch c.Op {
	case Eq, NotEq:
		if len(c.Args) != 1 {
			return false
		}
	case Gt, Gte, Lt, Lte:
		if len(c.Args) != 1 {
			return false
		}
	case In, NotIn:
		if len(c.Args) < 1 {
			return false
		}
	case Between:
		if len(c.Args) != 2 {
			return false
		}
	}

	return true
}

func (c AdvancedFilter) ToStatement(field string) (string, []any, bool) {
	if !c.validate() {
		return "", nil, false
	}

	var stmt string
	args := []any{}

	convertArgs := func() []any {
		for _, item := range c.Args {
			args = append(args, item)
		}

		return args
	}

	lookup := map[int]string{
		Eq:    "=",
		NotEq: "!=",
		Gt:    ">",
		Gte:   ">=",
		Lt:    "<",
		Lte:   "<=",
		In:    "IN",
		NotIn: "NOT IN",
	}

	switch c.Op {
	case Eq, NotEq, Gt, Gte, Lt, Lte:
		stmt = fmt.Sprintf("%v %v ?", field, lookup[c.Op])
		args = convertArgs()
	case In, NotIn:
		stmt = fmt.Sprintf("%v %v (%v)",
			field, lookup[c.Op], util.Float64ArrayToString(c.Args, ","),
		)
	case Between:
		stmt = fmt.Sprintf("%v BETWEEN ? AND ?", field)
		args = convertArgs()
	}

	return stmt, args, true
}
