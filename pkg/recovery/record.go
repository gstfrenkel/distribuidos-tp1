package recovery

import (
	"strconv"

	"tp1/pkg/amqp"
	"tp1/pkg/sequence"
)

type Record interface {
	toString() []string
	Header() amqp.Header
	SequenceIds() []sequence.Destination
	Message() []byte
}

type record struct {
	header        amqp.Header
	dstSequenceId []sequence.Destination
	msg           []byte
}

func NewRecord(header amqp.Header, dstSequenceId []sequence.Destination, msg []byte) Record {
	return &record{header: header, dstSequenceId: dstSequenceId, msg: msg}
}

func (r *record) toString() []string {
	recordStr := r.header.ToString()
	recordStr = append(recordStr, strconv.Itoa(len(r.dstSequenceId)))
	for _, id := range r.dstSequenceId {
		recordStr = append(recordStr, id.ToString())
	}
	recordStr = append(recordStr, string(r.msg))
	return recordStr
}

func (r *record) SequenceIds() []sequence.Destination {
	return r.dstSequenceId
}

func (r *record) Header() amqp.Header {
	return r.header
}

func (r *record) Message() []byte {
	return r.msg
}
