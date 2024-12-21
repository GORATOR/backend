package utils

import "strconv"

func StrToUint(str string) (uint, error) {
	result, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(result), nil
}
