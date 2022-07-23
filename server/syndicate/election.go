package syndicate

import "sync"

type ElectionSystem struct {
	sync.RWMutex
}
