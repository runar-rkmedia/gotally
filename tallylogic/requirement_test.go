package tallylogic

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStatsRequirement_Excludes(t *testing.T) {
	type fields struct {
		CellCount         *IntRequirement
		DuplicateFactors  *IntRequirement
		DuplicateValues   *IntRequirement
		UniqueFactorCount *IntRequirement
		WithValueCount    *IntRequirement
		UniquFactors      *IntListRequirement
		UniqeValues       *IntListRequirement
	}
	type args struct {
		stats GameStats
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StatsRequirement{
				CellCount:         tt.fields.CellCount,
				DuplicateFactors:  tt.fields.DuplicateFactors,
				DuplicateValues:   tt.fields.DuplicateValues,
				UniqueFactorCount: tt.fields.UniqueFactorCount,
				WithValueCount:    tt.fields.WithValueCount,
				UniquFactors:      tt.fields.UniquFactors,
				UniqeValues:       tt.fields.UniqeValues,
			}
			if got := s.Excludes(tt.args.stats); got != tt.want {
				t.Errorf("StatsRequirement.Excludes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntListRequirement_Excludes(t *testing.T) {
	type req struct {
		IncludesItems *[]uint64
		ExcludesItems *[]uint64
		OnlyItems     *[]uint64
	}
	type args struct {
	}
	tests := []struct {
		name string
		req  req
		list []uint64
		want bool
	}{
		{
			"Should include when req not set",
			req{},
			[]uint64{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := IntListRequirement{
				IncludesItems: tt.req.IncludesItems,
				ExcludesItems: tt.req.ExcludesItems,
				OnlyItems:     tt.req.OnlyItems,
			}
			if got := r.Excludes(tt.list); got != tt.want {
				t.Errorf("IntListRequirement.Excludes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntRequirement_Excludes(t *testing.T) {
	type req struct {
		GT  *int `json:"gt,omitempty"`
		GTE *int `json:"gte,omitempty"`
		EQ  *int `json:"eq,omitempty"`
		LT  *int `json:"lt,omitempty"`
		LTE *int `json:"lte,omitempty"`
	}
	type args struct {
		n int
	}
	tests := []struct {
		name string
		req  req
		n    int
		want bool
	}{
		// TODO: Add test cases.
		{
			"Should include when requirement is empty",
			req{},
			0,
			false,
		},
		{
			"Should exclude when requirement for gte",
			req{GTE: pint(5)},
			4,
			true,
		},
		{
			"Should include when requirement for gte",
			req{GTE: pint(4)},
			4,
			false,
		},
		{
			"Should include when requirement for gte and gt",
			req{GTE: pint(4), GT: pint(3)},
			4,
			false,
		},
		{
			"Should exclude when requirement for gte and gt",
			req{GTE: pint(4), GT: pint(4)},
			4,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := IntRequirement{
				GT:  tt.req.GT,
				GTE: tt.req.GTE,
				EQ:  tt.req.EQ,
				LT:  tt.req.LT,
				LTE: tt.req.LTE,
			}
			if got := r.Excludes(tt.n); got != tt.want {
				t.Errorf("%s IntRequirement.Excludes(%d) = %v, want %v", pretty(tt.req), tt.n, got, tt.want)
			}
		})
	}
}

func pretty(v any) string {
	b, _ := json.Marshal(v)
	return strings.ReplaceAll(string(b), `"`, "")
}

func pint(i int) *int {
	return &i
}
