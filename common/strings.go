package common

import "fmt"

func Inflect(n int, noun []string) string {
	var inflected string
	switch {
	case n%100 > 10 && n%100 < 15:
		inflected = noun[2]
	case n%10 == 1:
		inflected = noun[0]
	case n%10 > 1 && n%10 < 5:
		inflected = noun[1]
	default:
		inflected = noun[2]
	}

	return fmt.Sprintf("%d %s", n, inflected)
}
