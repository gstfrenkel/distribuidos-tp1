package percentile

import (
	"sort"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type filter struct {
	w             *worker.Worker
	n             uint8 //percentile value (0-100)
	scoredReviews map[message.GameId]message.ScoredReview
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	f.n = f.w.Query.(uint8)
	f.scoredReviews = make(map[message.GameId]message.ScoredReview)
	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	if messageId == message.EofMsg {
		f.publish()
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.saveScoredReview(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
	panic("implement me")
}

func (f *filter) saveScoredReview(msg message.ScoredReview) {
	review, exists := f.scoredReviews[msg.GameId]
	if !exists {
		f.scoredReviews[msg.GameId] = msg
	} else {
		review.Votes += msg.Votes
		f.scoredReviews[msg.GameId] = review
	}
}

func (f *filter) publish() {
	var topGames []string

	for _, reviews := range f.scoredReviews {
		if len(reviews) == 0 {
			continue
		}

		// Ordenar las reseñas por la cantidad de votos.
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].Votes < reviews[j].Votes
		})

		// Calcular el índice del percentil 90.
		index := int(float64(len(reviews)) * 0.9)
		if index >= len(reviews) {
			index = len(reviews) - 1
		}

		// Obtener el juego si está en el percentil 90 o superior.
		if len(reviews) > 0 && reviews[index].Votes > 0 {
			topGames = append(topGames, reviews[index].GameName)
		}
	}

	// Publicar el resultado (por ejemplo, imprimir o enviar por otro canal)
	logs.Logger.Infof("Top juegos dentro del percentil 90: %v", topGames)
}

func (f *filter) calculatePercentile() {
	for _, s := range f.scoredReviews {
		var votes []uint64
		for _, r := range s {
			votes = append(votes, r.Votes)
		}
		percentile := calculatePercentile(votes, float64(f.n))
	}

}

// calculatePercentile calculates the nth percentile of a slice of integers.
func calculatePercentile(data []int, percentile float64) int {
	if len(data) == 0 {
		return 0
	}

	sort.Ints(data)
	index := int(float64(len(data)-1) * percentile / 100.0)
	return data[index]
}
