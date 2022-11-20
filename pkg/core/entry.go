package core

import (
	"go.uber.org/zap"
)

var logger *zap.Logger = Logger

/****************
 ** Entry Type **
 ****************/

type EntryType int

const (
	OpenJob = iota
	CloseJob
)

// Convert an EntryType to a string
func (e EntryType) String() string {
	return [...]string{"OpenJob", "CloseJob"}[e]
}

/***********
 ** Entry **
 ***********/

// Entry is an entry in the log
type Entry struct {
	Type EntryType
	Term uint32
	Job  Job
}

/***************************
 ** Map LogEntry function **
 ***************************/

// Select range of log entries and return a new slice of log entries (start inclusive, end inclusive)
func ExtractListFromMap(m *map[uint32]Entry, start uint32, end uint32) []Entry {
	var list []Entry
	for i := start; i <= end; i++ {
		list = append(list, (*m)[i])
	}
	return list
}

// Flush the log entries after the given index. index + 1 and more are flushed but index is kept.
func FlushAfterIndex(m *map[uint32]Entry, index uint32) {
	for i := range *m {
		if i > index {
			delete(*m, i)
		}
	}
}
