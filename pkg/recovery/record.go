package recovery

import (
	"strconv"
	"tp1/pkg/sequence"

	"tp1/pkg/amqp"
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
	msg := r.header.ToString()
	msg = append(msg, strconv.Itoa(len(r.dstSequenceId)))
	for _, id := range r.dstSequenceId {
		msg = append(msg, id.ToString())
	}
	msg = append(msg, string(r.msg))
	return msg
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
