package development

import (
	"fmt"
)

func Solution(progresses []int, speeds []int) []int {

	remain := []int{}
	answers := []int{}

	for i, v := range progresses {
		remain = append(remain, (100-v)/speeds[i])
	}

	fmt.Println(remain)

	cnt, picked := 1, 0
	for i, v := range remain {
		if i == 0 {
			picked = remain[0]
		} else {
			if picked >= v {
				cnt++
			} else {
				answers = append(answers, cnt)
				picked = v
				cnt = 1
			}
		}
	}
	answers = append(answers, cnt)

	return answers
}
