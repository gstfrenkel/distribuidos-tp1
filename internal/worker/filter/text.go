package filter

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"

	"github.com/pemistahl/lingua-go"
)

var languages = map[string]lingua.Language{
	"arabic":     lingua.Arabic,
	"chinese":    lingua.Chinese,
	"english":    lingua.English,
	"french":     lingua.French,
	"german":     lingua.German,
	"italian":    lingua.Italian,
	"portuguese": lingua.Portuguese,
	"russian":    lingua.Russian,
	"spanish":    lingua.Spanish,
}

type text struct {
	w        *worker.Worker
	detector lingua.LanguageDetector
	target   lingua.Language
}

func NewText() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &text{w: w}, nil
}

func (f *text) Init() error {
	target, ok := languages[f.w.Query.(string)]
	if !ok {
		return errors.UnmappedLanguage
	}

	lang := make([]lingua.Language, 0, len(languages))
	for _, l := range languages {
		lang = append(lang, l)
	}

	f.detector = lingua.NewLanguageDetectorBuilder().FromLanguages(lang...).Build()
	f.target = target

	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *text) Start() {
	f.w.Start(f)
}

func (f *text) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var err error
	headers = headers.WithOriginId(amqp.ReviewOriginId)

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds, err = f.w.HandleEofMessage(delivery.Body, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	case message.ReviewWithTextID:
		msg, err := message.TextReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = f.publish(msg, headers)
		}
	default:
		logs.Logger.Infof(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *text) publish(msg message.TextReviews, headers amqp.Header) []sequence.Destination {
	sequenceIds := make([]sequence.Destination, 0, len(msg))

	for gameId, reviews := range msg {
		count := f.detectLang(reviews)

		if count == 0 {
			continue
		}

		k := shard.Int64(gameId, f.w.Outputs[0].Key, f.w.Outputs[0].Consumers)
		sequenceId := f.w.NextSequenceId(k)
		sequenceIds = append(sequenceIds, sequence.DstNew(k, sequenceId))
		b, err := message.ScoredReview{GameId: gameId, Votes: uint64(count)}.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		headers = headers.WithMessageId(message.ScoredReviewID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

		if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, k, b, headers); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}

	return sequenceIds
}

func (f *text) detectLang(reviews []string) int {
	count := 0
	for _, review := range reviews {
		lang, valid := f.detector.DetectLanguageOf(review)
		if valid && lang == f.target {
			count += 1
		}
	}
	return count
}

func (f *text) recover() {
	f.w.Recover(nil)
}
