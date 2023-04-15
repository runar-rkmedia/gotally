package tallylogiccompaction

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestUnmarshalCompactHistory(t *testing.T) {
	tests := []struct {
		name          string
		columns, rows int
		bytesAsHex    string
		want          string
		wantErr       bool
	}{
		// TODO: Add test cases.
		{
			"Single instruction swipe",
			5, 5,
			"7F",
			"D;",
			false,
		},
		{
			"Multiple Swipe instructions",
			5, 5,
			"7856",
			"D;L;U;R;",
			false,
		},
		{
			"Swipe with Path at end",
			5, 5,
			"741530",
			"D;indexes:21,20,15;",
			false,
		},
		{
			"Swipe with Path and then swipe",
			5, 5,
			"7415308F",
			"D;indexes:21,20,15;L;",
			false,
		},
		{
			"Path only",
			5, 5,
			"41530F",
			"indexes:21,20,15;",
			false,
		},
		{
			"Swipe 2x with Path at end",
			5, 5,
			"7841530F",
			"D;L;indexes:21,20,15;",
			false,
		},
		{
			"Swipe with longer Path at end",
			5, 5,
			"741530112F",
			"D;indexes:21,20,15,16,17,22;",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(tt.bytesAsHex)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%08b, %v", bytes, bytes)
			got, err := UnmarshalCompactHistory(bytes, tt.columns, tt.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalCompactHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotStr := got.DescribeShort()
			t.Logf("Got %s for input %x", gotStr, bytes)
			if !reflect.DeepEqual(gotStr, tt.want) {
				t.Log("got history raw", got)
				t.Errorf("UnmarshalCompactHistory() = %v, want %v", gotStr, tt.want)
			}
			// t.Fatal("Test completed")
		})
	}
}
