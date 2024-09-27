package review

import (
	"bytes"

	"tp1/pkg/ioutils"
	msg "tp1/pkg/message"
)

type message struct {
	appId       int64
	appName     string
	reviewText  string
	reviewScore int64
	reviewVotes int64
}

func New(appId int64, appName string, reviewText string, reviewScore int64, reviewVotes int64) msg.Message {
	return &message{
		appId:       appId,
		appName:     appName,
		reviewText:  reviewText,
		reviewScore: reviewScore,
		reviewVotes: reviewVotes,
	}
}

func FromBytes(b []byte) (msg.Message, error) {
	buf := bytes.NewBuffer(b)

	var msgId msg.ID
	var msgLen uint64
	fields := []interface{}{
		&msgId,
		&msgLen,
	}

	err := ioutils.ReadBytesFromBuff(fields, buf)
	if err != nil {
		return nil, err
	}

	appId, err := ioutils.ReadI64(buf)
	if err != nil {
		return nil, err
	}

	appNameLen, err := ioutils.ReadU64(buf)
	if err != nil {
		return nil, err
	}

	appName, err := ioutils.ReadString(buf, appNameLen)
	if err != nil {
		return nil, err
	}

	reviewTextLen, err := ioutils.ReadU64(buf)
	if err != nil {
		return nil, err
	}

	reviewText, err := ioutils.ReadString(buf, reviewTextLen)
	if err != nil {
		return nil, err
	}

	reviewScore, err := ioutils.ReadI64(buf)
	if err != nil {
		return nil, err
	}

	reviewVotes, err := ioutils.ReadI64(buf)
	if err != nil {
		return nil, err
	}

	return &message{
		appId:       appId,
		appName:     appName,
		reviewText:  reviewText,
		reviewScore: reviewScore,
		reviewVotes: reviewVotes,
	}, nil
}

func (m *message) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	fields := []interface{}{
		m.appId,
		uint64(len(m.appName)), []byte(m.appName),
		uint64(len(m.reviewText)), []byte(m.reviewText),
		m.reviewScore,
		m.reviewVotes,
	}

	err := ioutils.WriteBytesToBuff(fields, buf)
	if err != nil {
		return nil, err
	}

	msgLen := uint64(buf.Len())
	finalBuf := new(bytes.Buffer)

	fields = []interface{}{
		msg.ReviewIdMsg,
		msgLen,
		buf.Bytes(),
	}

	err = ioutils.WriteBytesToBuff(fields, finalBuf)
	if err != nil {
		return nil, err
	}

	return finalBuf.Bytes(), nil
}
