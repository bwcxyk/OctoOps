package scheduler

import (
	seatunnelModel "octoops/internal/model/seatunnel"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron
var taskEntryMap map[uint]cron.EntryID
var schedulerRunning atomic.Bool
var customTasks = map[uint]*CustomTask{}
var etlTasksMap = map[uint]*seatunnelModel.EtlTask{}

var mapsMu sync.RWMutex

type CustomTask struct {
	ID         uint          `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Spec       string        `json:"spec"`
	Status     int           `json:"status"`
	LastRun    time.Time     `json:"last_run"`
	NextRun    time.Time     `json:"next_run"`
	LastResult string        `json:"last_result"`
	EntryID    cron.EntryID  `json:"entry_id"`
	Job        func() string `json:"-"`
}

func computeNextRunFromEntry(entry cron.Entry, now time.Time) time.Time {
	if entry.Schedule != nil {
		return entry.Schedule.Next(now)
	}
	if !entry.Next.IsZero() {
		return entry.Next
	}
	return time.Time{}
}
