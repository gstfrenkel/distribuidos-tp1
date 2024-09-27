package message

import (
	"bytes"
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
		uint64(len(r.appName)), r.appName,
		uint64(len(r.reviewText)), r.reviewText,
		r.reviewScore,
		r.reviewVotes,
	}

	err := WriteBytesToBuff(fields, buf)
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

	err = WriteBytesToBuff(fields, finalBuf)
	if err != nil {
		return nil, err
	}

	return finalBuf.Bytes(), nil
}

func (r *ReviewMsg) FromBytes(b []byte) (*ReviewMsg, error) {
	buf := bytes.NewBuffer(b)

	var msgId messages.MessageId
	var msgLen uint64
	fields := []interface{}{
		&msgId,
		&msgLen,
	}

	err := ReadBytesFromBuff(fields, buf)
	if err != nil {
		return nil, err
	}

	appId, err := ReadI64(buf)
	if err != nil {
		return nil, err
	}

	appNameLen, err := ReadU64(buf)
	if err != nil {
		return nil, err
	}

	appName, err := ReadString(buf, appNameLen)
	if err != nil {
		return nil, err
	}

	reviewTextLen, err := ReadU64(buf)
	if err != nil {
		return nil, err
	}

	reviewText, err := ReadString(buf, reviewTextLen)
	if err != nil {
		return nil, err
	}

	reviewScore, err := ReadI64(buf)
	if err != nil {
		return nil, err
	}

	reviewVotes, err := ReadI64(buf)
	if err != nil {
		return nil, err
	}

	return &ReviewMsg{
		appId:       appId,
		appName:     appName,
		reviewText:  reviewText,
		reviewScore: reviewScore,
		reviewVotes: reviewVotes,
	}, nil
}
