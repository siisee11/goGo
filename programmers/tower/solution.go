package tower

type tower struct {
	idx    int
	height int
}

func Solution(heights []int) []int {
	dp := []tower{}
	answer := []int{}

	for i, v := range heights {
		if i == 0 {
			dp = append(dp, tower{0, 0})
			answer = append(answer, 0)
			continue
		}
		if heights[i-1] > v {
			dp = append(dp, tower{i, heights[i-1]})
			answer = append(answer, i)
		} else {
			var j int
			for j = i - 1; j >= 0; j-- {
				if dp[j].height > v {
					j--
					break
				}
			}
			j++
			dp = append(dp, tower{dp[j].idx, dp[j].height})
			answer = append(answer, dp[j].idx)
		}
	}

	return answer
}
