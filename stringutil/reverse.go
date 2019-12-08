// 패키지 stringutil 은 문자열 작업을 위한 유틸리티 함수들을 포함하고 있음
package stringutil

// Reverse 는 인자 문자열을 rune-wise 방식으로 왼쪽에서 오른쪽으로 반전하여 리턴합니다.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
