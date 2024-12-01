package aggregator

import (
	"tp1/internal/errors"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/utils/shard"
)

func shardOutput(output amqp.Destination, clientId string) amqp.Destination {
	output, err := shard.AggregatorOutput(output, clientId)

	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}
	return output
}
