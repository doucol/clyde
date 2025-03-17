package util

import "errors"

func IsErr(this error, those ...error) bool {
	for _, err := range those {
		if errors.Is(this, err) {
			return true
		}
	}
	return false
}
