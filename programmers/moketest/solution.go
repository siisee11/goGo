package moketest

func Solution(answers []int) []int {
	answer := []int{}

	picker1 := [5]int{1, 2, 3, 4, 5}
	picker2 := [8]int{2, 1, 2, 3, 2, 4, 2, 5}
	picker3 := [10]int{3, 3, 1, 1, 2, 2, 4, 4, 5, 5} 

	correct := [3]int{0, 0, 0}
	for i, v := range answers {
		if v == picker1[i % len(picker1)] {
			correct[0]++
		}
		if v == picker2[i % len(picker2)] {
			correct[1]++
		}
		if v == picker3[i % len(picker3)] {
			correct[2]++
		}	
	}

	m := 0
	for i, e := range correct {
		if i==0 || e > m {
			m = e
		}
	}

	for i, e := range correct {
		if e == m {
			answer = append(answer, i + 1)
		}
	}

	return answer
}
