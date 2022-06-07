package benchmark

import (
	"fmt"
	"server/gamelog"
	"time"
)

type Record struct {
	key       string
	startedAt time.Time
	endedAt   time.Time
}

type BenchMarker struct {
	RecordMap map[string]*Record
}

// New create a new benchmark instance
func New() *BenchMarker {
	bm := &BenchMarker{
		RecordMap: make(map[string]*Record),
	}

	return bm
}

func (bm *BenchMarker) Start(key string) {
	// check record exists
	if _, ok := bm.RecordMap[key]; ok {
		gamelog.L.Warn().Msgf(`Benchmark on key "%s" has been override`, key)
		return
	}

	// store record
	bm.RecordMap[key] = &Record{
		key:       key,
		startedAt: time.Now(),
	}
}

func (bm *BenchMarker) End(key string) {
	now := time.Now()

	r, ok := bm.RecordMap[key]
	if !ok {
		gamelog.L.Warn().Msgf(`Benchmark on key "%s" does not exists`, key)
		return
	}

	r.endedAt = now
}

func (bm *BenchMarker) ReportGet() (time.Duration, []string, error) {
	if len(bm.RecordMap) == 0 {
		gamelog.L.Warn().Msg("There is no benchmark record")
		return 0, nil, fmt.Errorf("benchmark record not found")
	}

	// calculate duration
	var totalTime time.Duration
	var reports []string

	for key, record := range bm.RecordMap {
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

func (bm *BenchMarker) Report() {
	if len(bm.RecordMap) == 0 {
		gamelog.L.Warn().Msg("There is no benchmark record to report")
		return
	}

	// calculate duration
	totalTime, reports, err := bm.ReportGet()
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("Failed to get benchmark report")
		return
	}

	// log report if alert baseline is not set or total time exceed alert baseline
	gamelog.L.Info().Interface("report", reports).Int64("total ms", totalTime.Milliseconds()).Msg("Benchmark report")
}

func (bm *BenchMarker) Alert(millisecond int64) {
	if len(bm.RecordMap) == 0 {
		gamelog.L.Warn().Msg("There is no benchmark record to alert")
		return
	}

	totalTime, reports, err := bm.ReportGet()
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("Failed to get benchmark report")
		return
	}

	if totalTime.Milliseconds() > millisecond {
		gamelog.L.Warn().Interface("report", reports).Int64("total ms", totalTime.Milliseconds()).Int64("required ms", millisecond).Msg("Exceed required time")
	}
}
