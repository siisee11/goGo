package tower

import (
	"reflect"
	"testing"
)

func TestSolution(t *testing.T) {
	cases := []struct {
		in1, want []int
	}{
		{
			in1:  []int{6, 9, 5, 7, 4},
			want: []int{0, 0, 2, 2, 4},
		},
		{
			in1:  []int{1, 2, 3, 4, 5, 6, 7},
			want: []int{0, 0, 0, 0, 0, 0, 0},
		},
		{
			in1: []int{7, 6, 5, 4, 3, 2, 1},
			want: []int{0, 1, 2, 3, 4, 5, 6},
		},
		{
			in1: []int{9, 1, 1, 1},
			want: []int{0, 1, 1, 1},
		},
	}
	for _, c := range cases {
		got := Solution(c.in1)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("\nSolution(%v) == %v, want %v", c.in1, got, c.want)
		}
	}
}
