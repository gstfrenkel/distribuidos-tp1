package shard

import (
	"fmt"
	"strconv"

	"tp1/pkg/amqp"
	"tp1/pkg/utils/id"

	"github.com/pierrec/xxHash/xxHash32"
)

// String shards an id into a value between 0 and consumers, and inserts it into the key.
func String(id string, key string, consumers uint8) string {
	if consumers == 0 {
		return key
	}
	return fmt.Sprintf(key, xxHash32.Checksum([]byte(id), 0)%uint32(consumers))
}

// Int64 shards an id into a value between 0 and consumers, and inserts it into the key.
func Int64(id int64, key string, consumers uint8) string {
	if consumers == 0 {
		return key
	}
	return fmt.Sprintf(key, xxHash32.Checksum([]byte{byte(id)}, 0)%uint32(consumers))
}

// AggregatorOutput returns an output with an updated and ready-to-use key.
func AggregatorOutput(output amqp.Destination, clientId string) (amqp.Destination, error) {
	parts, err := id.SplitId(clientId)
	if err != nil {
		return output, err
	}
	gatewayId, _ := strconv.Atoi(parts[0])
	output.Key = fmt.Sprintf(output.Key, gatewayId)
	return output, nil
}
