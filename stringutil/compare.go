package stringutil

type Diff int

const (
	Less    Diff = -1
	Equal   Diff = 0
	Greater Diff = 1
)

func Compare(str1, str2 string) Diff {
	if str1 == str2 {
		return Equal
	}

	raw1 := []rune(str1)
	raw2 := []rune(str2)

	min := len(raw1)
	if len(raw2) < min {
		min = len(raw2)
	}

	for k := 0; k < min; k++ {
		if raw1[k] < raw2[k] {
			return Less

		} else if raw1[k] > raw2[k] {
			return Greater
		}
	}

	if len(raw1) < len(raw2) {
		return Less
	}
	return Greater
}
