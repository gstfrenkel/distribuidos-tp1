package message

import (
	"bytes"
	"tp1/pkg/utils/encoding"
)

type Review []review

type review struct {
	GameId   int64
	GameName string
	Text     string
	Score    int8
}

func ReviewFromBytes(b []byte) (Review, error) {
	buf := bytes.NewBuffer(b)

	size, err := encoding.DecodeUint32(buf)
	if err != nil {
		return nil, err
	}

	reviews := make([]review, 0, size)
	for i := uint32(0); i < size; i++ {
		rev, err := reviewFromBuf(buf)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}

	return reviews, nil
}

func reviewFromBuf(buf *bytes.Buffer) (review, error) {
	id, err := encoding.DecodeInt64(buf)
	if err != nil {
		return review{}, err
	}

	name, err := encoding.DecodeString(buf)
	if err != nil {
		return review{}, err
	}

	text, err := encoding.DecodeString(buf)
	if err != nil {
		return review{}, err
	}

	score, err := encoding.DecodeInt8(buf)
	if err != nil {
		return review{}, err
	}

	return review{
		GameId:   id,
		GameName: name,
		Text:     text,
		Score:    score,
	}, nil
}

func ReviewsFromClientReviews(clientReview []DataCSVReviews) ([]byte, error) {
	rs := make(Review, 0, len(clientReview))
	for _, r := range clientReview {
		rs = append(rs, review{
			GameId:   r.AppID,
			GameName: r.AppName,
			Text:     r.ReviewText,
			Score:    int8(r.ReviewScore),
		})
	}

	return rs.ToBytes()
}

func (m Review) ToBytes() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	if err := encoding.EncodeNumber(buf, uint32(len(m))); err != nil {
		return nil, err
	}

	for _, rev := range m {
		if err := encoding.EncodeNumber(buf, rev.GameId); err != nil {
			return nil, err
		}

		if err := encoding.EncodeString(buf, rev.GameName); err != nil {
			return nil, err
		}

		if err := encoding.EncodeString(buf, rev.Text); err != nil {
			return nil, err
		}

		if err := encoding.EncodeNumber(buf, rev.Score); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (m Review) ToScoredReviewMessage(targetScore int8) ScoredReviews {
	scoredReviewMsg := ScoredReviews{}
	gameVotesMap := map[int64]int64{}
	gameNamesMap := map[int64]string{}

	for _, reviewMsg := range m {
		if reviewMsg.Score != targetScore {
			continue
		}

		if count, exists := gameVotesMap[reviewMsg.GameId]; exists {
			gameVotesMap[reviewMsg.GameId] = count + 1
		} else {
			gameVotesMap[reviewMsg.GameId] = 1
			gameNamesMap[reviewMsg.GameId] = reviewMsg.GameName
		}
	}

	for k, v := range gameVotesMap {
		scoredReviewMsg = append(scoredReviewMsg, ScoredReview{GameId: k, GameName: gameNamesMap[k], Votes: uint64(v)})
	}

	return scoredReviewMsg
}

func (m Review) ToReviewWithTextMessage(targetScore int8) TextReviews {
	textReviewMessage := TextReviews{}

	for _, reviewMsg := range m {
		if reviewMsg.Score != targetScore {
			continue
		}

		if _, exists := textReviewMessage[reviewMsg.GameId]; exists {
			textReviewMessage[reviewMsg.GameId] = append(textReviewMessage[reviewMsg.GameId], reviewMsg.Text)
		} else {
			textReviewMessage[reviewMsg.GameId] = []string{reviewMsg.Text}
		}
	}

	return textReviewMessage
}
