package gateway

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
	"tp1/pkg/message"
)

func TestReadPayloadSize_ValidInput(t *testing.T) {
	data := make([]byte, 9)
	expectedPayloadSize := uint64(1)
	binary.BigEndian.PutUint64(data, expectedPayloadSize)
	read := 9
	var payloadSize uint64
	remaining := readPayloadSize(&payloadSize, data, &read)
	assert.Equal(t, expectedPayloadSize, payloadSize)
	assert.Equal(t, 1, read)
	assert.Equal(t, []byte{0}, remaining)
}

func TestReadId_ValidInput(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	read := 3
	var msgId uint8
	remaining := readId(&msgId, data, &read)
	assert.Equal(t, uint8(0x01), msgId)
	assert.Equal(t, 2, read)
	assert.Equal(t, []byte{0x02, 0x03}, remaining)
}

func TestHasReadCompletePayload_True(t *testing.T) {
	read := 10
	payloadSize := uint64(10)
	assert.True(t, hasReadCompletePayload(read, payloadSize))
}

func TestHasReadCompletePayload_False(t *testing.T) {
	read := 5
	payloadSize := uint64(10)
	assert.False(t, hasReadCompletePayload(read, payloadSize))
}

func TestHasReadPayloadSize_True(t *testing.T) {
	read := 8
	payloadSize := uint64(0)
	assert.True(t, hasReadPayloadSize(read, payloadSize))
}

func TestHasReadPayloadSize_False(t *testing.T) {
	read := 5
	payloadSize := uint64(0)
	assert.False(t, hasReadPayloadSize(read, payloadSize))
}

func TestHasReadId_True(t *testing.T) {
	read := 1
	msgId := uint8(0)
	assert.True(t, hasReadId(read, msgId))
}

func TestHasReadId_False(t *testing.T) {
	read := 0
	msgId := uint8(0)
	assert.False(t, hasReadId(read, msgId))
}

func TestProcessPayload_EndOfFile(t *testing.T) {
	g := &Gateway{ChunkChan: make(chan ChunkItem, 1)}
	msgId := message.ReviewIdMsg
	var payload []byte
	payloadLen := uint64(0)
	eofs := uint8(0)

	err, remaining := processPayload(g, msgId, payload, payloadLen, &eofs)
	assert.Nil(t, err)
	assert.Equal(t, payload, remaining)
	assert.Equal(t, uint8(1), eofs)
	assert.Equal(t, <-g.ChunkChan, ChunkItem{MsgId: msgId, Msg: nil})
}

func TestProcessPayload_ValidPayload(t *testing.T) {
	g := &Gateway{ChunkChan: make(chan ChunkItem, 1)}
	msgId := message.ReviewIdMsg
	csvReviews := message.DataCSVReviews{
		AppID:       1,
		AppName:     "juego",
		ReviewText:  "buenisimo",
		ReviewScore: 1,
		ReviewVotes: 0,
	}
	payload, _ := csvReviews.ToBytes()

	payloadLen := uint64(len(payload))
	eofs := uint8(0)

	err, remaining := processPayload(g, msgId, payload, payloadLen, &eofs)
	assert.Nil(t, err)
	assert.Equal(t, payload, remaining)
	assert.Equal(t, <-g.ChunkChan, ChunkItem{MsgId: msgId, Msg: csvReviews})
}
