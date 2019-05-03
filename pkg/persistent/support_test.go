package persistent

import (
	"fmt"
	"time"
)

var _ Keyer = staticKeyer("")

type staticKeyer string

func (s staticKeyer) Key(key string) (string, time.Duration) {
	return fmt.Sprintf("%s/%s", s, key), time.Duration(0)
}
