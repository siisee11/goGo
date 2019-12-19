package development

import (
	"reflect"
	"testing"
)

func TestSolution(t *testing.T) {
	cases := []struct {
		in1, in2, want []int
	}{
		{
			in1:  []int{93, 30, 55},
			in2:  []int{1, 30, 5},
			want: []int{2, 1},
		},
	}
	for _, c := range cases {
		got := Solution(c.in1, c.in2)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("\nSolution(%v, %v) == %v, want %v", c.in1, c.in2, got, c.want)
		}
	}
}
