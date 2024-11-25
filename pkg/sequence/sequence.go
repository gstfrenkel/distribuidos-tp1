package sequence

import "fmt"

const (
	partsAmount = 2
	separator   = "-"
)

func errInvalidSequence(seq string) error {
	return fmt.Errorf("invalid sequence: %s", seq)
}
