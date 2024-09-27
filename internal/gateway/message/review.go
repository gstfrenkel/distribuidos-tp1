package message

import (
	"bytes"
	"tp1/pkg/ioutils"
	"tp1/pkg/messages"
)

type ReviewMsg struct {
	appId       int64
	appName     string
	reviewText  string
	reviewScore int64
	reviewVotes int64
}

func (r *ReviewMsg) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	fields := []interface{}{
		r.appId,
		uint64(len(r.appName)), []byte(r.appName),
		uint64(len(r.reviewText)), []byte(r.reviewText),
		r.reviewScore,
		r.reviewVotes,
	}

	err := ioutils.WriteBytesToBuff(fields, buf)
	if err != nil {
		return nil, err
	}

	msgLen := uint64(buf.Len())
	finalBuf := new(bytes.Buffer)

	fields = []interface{}{
		messages.REVIEW_ID_MSG,
		msgLen,
		buf.Bytes(),
	}

	err = ioutils.WriteBytesToBuff(fields, finalBuf)
	if err != nil {
		return nil, err
	}

	return finalBuf.Bytes(), nil
}

func FromBytes(b []byte) error {
	buf := bytes.NewBuffer(b)

	var msgId messages.MessageId
	var msgLen uint64
	fields := []interface{}{
		msgId,
		msgLen,
	}

	err := ioutils.ReadBytesFromBuff(fields, buf)
	if err != nil {
		return err
	}

	appId, err := ioutils.ReadI64(buf)
	if err != nil {
		return err
	}

	appNameLen, err := ioutils.ReadU64(buf)
	if err != nil {
		return err
	}

	appName, err := ioutils.ReadString(buf, appNameLen)
	if err != nil {
		return err
	}

	reviewTextLen, err := ioutils.ReadU64(buf)
	if err != nil {
		return err
	}

	reviewText, err := ioutils.ReadString(buf, reviewTextLen)
	if err != nil {
		return err
	}

	reviewScore, err := ioutils.ReadI64(buf)
	if err != nil {
		return err
	}

	reviewVotes, err := ioutils.ReadI64(buf)
	if err != nil {
		return err
	}

	return &ReviewMsg{
		appId:       appId,
		appName:     appName,
		reviewText:  reviewText,
		reviewScore: reviewScore,
		reviewVotes: reviewVotes,
	}, nil
}
