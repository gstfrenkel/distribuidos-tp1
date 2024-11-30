package shard

import (
	"fmt"

	"github.com/pierrec/xxHash/xxHash32"
)

func String(id string, key string, consumers uint8) string {
	if consumers == 0 {
		return key
	}
	return fmt.Sprintf(key, xxHash32.Checksum([]byte(id), 0)%uint32(consumers))
}

func Int64(id int64, key string, consumers uint8) string {
	if consumers == 0 {
		return key
	}
	return fmt.Sprintf(key, xxHash32.Checksum([]byte{byte(id)}, 0)%uint32(consumers))
}
