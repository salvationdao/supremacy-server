package benchmark

import (
	"fmt"
	"golang.org/x/exp/slices"
	"server/gamelog"
	"time"
)

type Record struct {
	key       string
	startedAt time.Time
	endedAt   time.Time
}

type Benchmarker struct {
	RecordList []*Record
}

// New create a new benchmark instance
func New() *Benchmarker {
	bm := &Benchmarker{
		RecordList: []*Record{},
	}

	return bm
}

func (bm *Benchmarker) Start(key string) {
	// check record exists
	index := slices.IndexFunc(bm.RecordList, func(r *Record) bool { return r.key == key })
	if index != -1 {
		gamelog.L.Warn().Msgf(`Benchmark on key "%s" has been override`, key)
		return
	}

	// store record
	bm.RecordList = append(bm.RecordList, &Record{
		key:       key,
		startedAt: time.Now(),
	})
}

func (bm *Benchmarker) End(key string) {
	now := time.Now()

	index := slices.IndexFunc(bm.RecordList, func(r *Record) bool { return r.key == key })

	if index == -1 {
		gamelog.L.Warn().Msgf(`Benchmark on key "%s" does not exists`, key)
		return
	}

	bm.RecordList[index].endedAt = now
}

func (bm *Benchmarker) ReportGet() (time.Duration, []string, error) {
	if len(bm.RecordList) == 0 {
		gamelog.L.Debug().Msg("There is no benchmark record")
		return 0, nil, fmt.Errorf("benchmark record not found")
	}

	// calculate duration
	var totalTime time.Duration
	var reports []string

	for key, record := range bm.RecordList {
		if record.startedAt.After(record.endedAt) {
			gamelog.L.Warn().Msgf(`The end time of key "%s" is not set`, key)
			return 0, nil, fmt.Errorf("invalid benchmark record")
		}
		duration := record.endedAt.Sub(record.startedAt)
		reports = append(reports, fmt.Sprintf(`%s: %d ms`, key, duration.Milliseconds()))

		totalTime += duration
	}

	return totalTime, reports, nil
}

func (bm *Benchmarker) Alert(millisecond int64) {
	if len(bm.RecordList) == 0 {
		gamelog.L.Debug().Msg("There is no benchmark record to alert")
		return
	}

	totalTime, reports, err := bm.ReportGet()
	if err != nil {
		gamelog.L.Debug().Err(err).Msg("Failed to get benchmark report")
		return
	}

	if totalTime.Milliseconds() > millisecond {
		gamelog.L.Warn().Interface("report", reports).Int64("total ms", totalTime.Milliseconds()).Int64("required ms", millisecond).Msg("Exceed required time")
	}

	// free up memory
	bm.RecordList = nil
}
