package text

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"

	"github.com/pemistahl/lingua-go"
)

var (
	headersEof = map[string]any{amqp.OriginIdHeader: amqp.ReviewOriginId, amqp.MessageIdHeader: uint8(message.EofMsg)}
	headers    = map[string]any{amqp.MessageIdHeader: uint8(message.ScoredReviewID)}
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

type filter struct {
	w        *worker.Worker
	detector lingua.LanguageDetector
	target   lingua.Language
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
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

	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		headersEof[amqp.ClientIdHeader] = delivery.Headers[amqp.ClientIdHeader]
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, headersEof, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	} else if messageId == message.ReviewWithTextID {
		msg, err := message.TextReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		headers[amqp.ClientIdHeader] = delivery.Headers[amqp.ClientIdHeader]
		f.publish(msg)
	} else {
		logs.Logger.Infof(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(msg message.TextReviews) {
	for gameId, reviews := range msg {
		count := 0

		for _, review := range reviews {
			lang, valid := f.detector.DetectLanguageOf(review)
			if valid && lang == f.target {
				count += 1
			}
		}

		if count == 0 {
			continue
		}

		k := worker.ShardGameId(gameId, f.w.Outputs[0].Key, f.w.Outputs[0].Consumers)
		b, err := message.ScoredReview{GameId: gameId, Votes: uint64(count)}.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, k, b, headers); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}
