package kth

import "testing"

func TestSolution(t *testing.T) {

	got := Solution([1, 5, 2, 6 ,3, 7, 4], [[2, 5, 3], [4, 4, 1], [1, 7, 3]])

	if got != [5, 6, 3] {
		t.Errorf("Solution() == %q, want %q", got, [5, 6, 3])
	}

}

	/*
	cases := []struct {
		in1, want int[]
		in2 int[][]
	}{
		{
			in1 : [1, 5, 2, 6 ,3, 7, 4], 
			in2 : [[2, 5, 3], [4, 4, 1], [1, 7, 3]], 
			want : [5, 6, 3]
		},
	}
	for _, c := range cases {
		got := Solution(c.in)
		if got != c.want {
			t.Errorf("Solution(%q) == %q, want %q", c.in, got, c.want)
		}
	}
	*/