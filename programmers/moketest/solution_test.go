package moketest

import (
	"reflect"
	"testing"
)

func TestSolution(t *testing.T) {
	cases := []struct {
		in1, want []int
	}{
		{
			in1:  []int{1, 2, 3, 4, 5},
			want: []int{1},
		},
		{
			in1: []int{1, 3, 2, 4, 2},
			want: []int{1, 2, 3},
		},
	}
	for _, c := range cases {
		got := Solution(c.in1)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("\nSolution(%v) == %v, want %v", c.in1, got, c.want)
		}
	}
}