package persistent

import "fmt"

type staticKeyer string

func (s staticKeyer) Key(key string) string {
	return fmt.Sprintf("%s/%s", s, key)
}
