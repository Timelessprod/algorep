package core

import (
	"fmt"
	"io/ioutil"
)

const NO_WORKER = -1

/****************
 ** Job Status **
 ****************/

type JobState int

const (
	JobWaiting = iota
	JobDone
)

// Convert a JobStatus to a string
func (s JobState) String() string {
	return [...]string{"WAITING", "DONE"}[s]
}

/*********
 ** Job **
 *********/

type Job struct {
	Id       uint32
	Term     uint32
	State    JobState
	WorkerId int
	Input    string
	Output   string
}

// Get the reference `Id-Term` of the job
func (job *Job) GetReference() string {
	return fmt.Sprintf("%d-%d", job.Id, job.Term)
}

/***************
 ** Load Code **
 ***************/

// Load code from a file
func LoadCodeFromFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	return string(content), err
}
