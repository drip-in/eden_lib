package pubsub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type Test struct {
	name   string
	method string
	run    func()
}

var Tests []Test

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type QueueTestSuite struct {
	suite.Suite
	conf AsyncQueueConfig
}

func (suite *QueueTestSuite) SetupTest() {
}

func (suite *QueueTestSuite) TestEnqueueSampleJob() {
	job := NewSampleJob("449", 0)
	queue := NewAsyncQueue(suite.conf)
	err := queue.Enqueue(job)
	suite.NoError(err, "Expected no error when enqueuing job")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPlaceOrderJobHandler(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}

func NewSampleJob(id string, st int64) SampleJob[SampleJobArg] {
	return SampleJob[SampleJobArg]{
		Argument: SampleJobArg{StartTime: st},
		ID:       id,
	}
}

type SampleJob[T SampleJobArg] struct {
	Argument SampleJobArg
	ID       string
}

type SampleJobArg struct {
	StartTime int64
}

func (s SampleJobArg) JSON() ([]byte, error) {
	return json.Marshal(s)
}

func (j SampleJob[SampleJobArg]) Name() string {
	return "sample_job"
}

func (j SampleJob[SampleJobArg]) Arg() JobArg {
	return j.Argument
}

func (j SampleJob[SampleJobArg]) Config() JobConfig {
	retry := 5
	return JobConfig{
		Queue:   Critical,
		Retry:   &retry,
		JobID:   j.ID,
		Timeout: 50 * time.Second,
		Unique:  5 * time.Second,
	}
}
