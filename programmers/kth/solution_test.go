package kth

import (
	"reflect"
	"testing"
)

func TestSolution(t *testing.T) {
	cases := []struct {
		in1, want []int
		in2       [][]int
	}{
		{
			in1:  []int{1, 5, 2, 6, 3, 7, 4},
			in2:  [][]int{{2, 5, 3}, {4, 4, 1}, {1, 7, 3}},
			want: []int{5, 6, 3},
		},
	}
	for _, c := range cases {
		got := Solution(c.in1, c.in2)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("\nSolution(%v, %v) == %v, want %v", c.in1, c.in2, got, c.want)
		}
	}
}
