package tallylogic

import (
	"fmt"
	"testing"
)

func TestCell_Value(t *testing.T) {
	tests := []struct {
		baseValue int64
		power     int
		want      int64
	}{
		{0, 0, 0},
		{1, 0, 1},
		{1, 1, 2},
		{1, 2, 4},
		{1, 3, 8},
		{1, 4, 16},
		{2, 0, 2},
		{2, 1, 4},
		{2, 2, 8},
		{2, 3, 16},
		{2, 4, 32},
		{7, 0, 7},
		{7, 1, 14},
		{7, 2, 28},
		{7, 3, 56},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("should return correct for %d %d = %d", tt.baseValue, tt.power, tt.want), func(t *testing.T) {
			c := Cell{
				baseValue: tt.baseValue,
				power:     tt.power,
			}
			if got := c.Value(); got != tt.want {
				t.Errorf("Cell.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
