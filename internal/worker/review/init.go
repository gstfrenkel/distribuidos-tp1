package review

import (
	"fmt"

	"tp1/pkg/broker"
)

func (f Filter) queues() []string {
	q := []string{f.config.String("positive-reviews.queue-name", "positive_reviews")}

	consumers := f.config.Int("positive-reviews-sh.consumers", 1)
	qName := f.config.String("positive-reviews-sh.queue-name", "positive_reviews_%d")
	for i := 1; i <= consumers; i++ {
		q = append(q, fmt.Sprintf(qName, i))
	}

	consumers = f.config.Int("negative-reviews-sh.consumers", 1)
	qName = f.config.String("negative-reviews-sh.queue-name", "negative_reviews_%d")
	for i := 1; i <= consumers; i++ {
		q = append(q, fmt.Sprintf(qName, i))
	}

	return q
}

func (f Filter) binds() []broker.QueueBind {
	ex := f.config.String("outputExchange.publishing-name", "reviews")

	b := []broker.QueueBind{{
		Name:     f.config.String("positive-reviews.queue-name", "positive_reviews"),
		Key:      "",
		Exchange: ex,
	}}

	consumers := f.config.Int("positive-reviews-sh.consumers", 1)
	qName := f.config.String("positive-reviews-sh.queue-name", "positive_reviews_%d")
	for i := 1; i <= consumers; i++ {
		b = append(b, broker.QueueBind{
			Name:     fmt.Sprintf(qName, i),
			Key:      fmt.Sprintf(f.config.String("positive-reviews-sh.key", "p%d"), i-1),
			Exchange: ex,
		})
	}

	consumers = f.config.Int("negative-reviews-sh.consumers", 1)
	qName = f.config.String("negative-reviews-sh.queue-name", "negative_reviews_%d")
	for i := 1; i <= consumers; i++ {
		b = append(b, broker.QueueBind{
			Name:     fmt.Sprintf(qName, i),
			Key:      fmt.Sprintf(f.config.String("negative-reviews-sh.key", "n%d"), i-1),
			Exchange: ex,
		})
	}

	return b
}
