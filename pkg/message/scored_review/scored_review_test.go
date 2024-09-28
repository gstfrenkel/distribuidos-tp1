package scored_review

import (
	"testing"
	"tp1/pkg/message/utils"

	"github.com/stretchr/testify/assert"
)

func TestToBytes(t *testing.T) {
	original := New(map[utils.Key]int64{
		utils.Key{GameId: 1, GameName: "Game1"}: 4,
		utils.Key{GameId: 2, GameName: "Game1"}: 1,
		utils.Key{GameId: 3, GameName: "Game1"}: 8,
	})

	serialized, err := original.ToBytes()

	assert.NoError(t, err)
	assert.NotEmpty(t, serialized)
}

// Test for FromBytes method
func TestFromBytes(t *testing.T) {
	original := New(map[utils.Key]int64{
		utils.Key{GameId: 1, GameName: "Game1"}: 4,
		utils.Key{GameId: 2, GameName: "Game1"}: 1,
		utils.Key{GameId: 3, GameName: "Game1"}: 8,
	})

	serialized, err := original.ToBytes()
	assert.NoError(t, err)

	deserialized, err := FromBytes(serialized)
	assert.NoError(t, err)

	assert.Equal(t, original, deserialized)
}
