package worker

import "fmt"

const NO_WORKER = -1

/****************
 ** Job Status **
 ****************/

type JobStatus int

const (
	JobWaiting JobStatus = iota
	JobDone    JobStatus = iota
)

// Convert a JobStatus to a string
func (s JobStatus) String() string {
	return [...]string{"Waiting", "Done"}[s]
}

/*********
 ** Job **
 *********/

type Job struct {
	Id     uint32
	Term   uint32
	Status JobStatus
	Input  string
	Output string
}

// Get the reference `Id-Term` of the job
func (job *Job) GetReference() string {
	return fmt.Sprintf("%d-%d", job.Id, job.Term)
}
