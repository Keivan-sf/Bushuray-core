package cmd

import "sync"

var ConnectionMutex = sync.Mutex{}
