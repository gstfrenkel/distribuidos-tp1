package recovery

import (
	"strconv"
	"tp1/pkg/sequence"

	"tp1/pkg/amqp"
)

type Record interface {
	toString() []string
}

type record struct {
	header        amqp.Header
	dstSequenceId []sequence.Destination
	msg           []string
}

func NewRecord(header amqp.Header, dstSequenceId []sequence.Destination, msg []string) Record {
	return &record{header: header, dstSequenceId: dstSequenceId, msg: msg}
}

func (r *record) toString() []string {
	msg := r.header.ToString()
	msg = append(msg, strconv.Itoa(len(r.dstSequenceId)))
	for _, id := range r.dstSequenceId {
		msg = append(msg, id.ToString())
	}
	msg = append(msg, r.msg...)
	return msg
}
