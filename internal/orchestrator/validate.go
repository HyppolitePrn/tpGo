package orchestrator

import "fmt"

func ValidateWorkers(n int) (int, error) {
	if n < 1 || n > 100 {
		return 3, fmt.Errorf("invalid worker count %d: must be between 1 and 100", n)
	}
	return n, nil
}
