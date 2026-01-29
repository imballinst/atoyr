package utils

import "time"

func ScrambleWord(word string) string {
	runes := []rune(word)
	n := len(runes)

	for i := n - 1; i > 0; i-- {
		j := time.Now().UnixNano() % int64(i+1)
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
