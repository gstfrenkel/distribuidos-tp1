package percentile

import (
	"tp1/internal/errors"
	"tp1/pkg/logs"
)

func SendBatches(slice []any, batchSize uint16, toBytes func([]any) ([]byte, error), sendBatch func([]byte, string), clientId string) {
	length := len(slice)
	for start := 0; start < length; {
		batch, nextStart := nextBatch(slice, length, batchSize, start)
		bytes, err := toBytes(batch)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err)
			return
		}
		sendBatch(bytes, clientId)
		start = nextStart
	}
}

func nextBatch(data []any, lenGames int, batchSize uint16, start int) ([]any, int) {
	end := start + int(batchSize)
	if end > lenGames {
		end = lenGames
	}
	return data[start:end], end
}
