package gateway

import (
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type ChunkSender struct {
	id           int //0 for games, 1 for reviews
	channel      <-chan ChunkItem
	broker       amqp.MessageBroker
	exchange     string
	chunk        []any
	maxChunkSize uint8
	routingKey   string
}

type ChunkItem struct {
	MsgId message.ID
	Msg   any //DataCSVGames or DataCSVReviews
}

func newChunkSender(id int, channel <-chan ChunkItem, broker amqp.MessageBroker, exchange string, chunkMaxSize uint8, routingKey string) *ChunkSender {
	return &ChunkSender{
		id:           id,
		channel:      channel,
		broker:       broker,
		exchange:     exchange,
		chunk:        make([]any, 0, chunkMaxSize),
		maxChunkSize: chunkMaxSize,
		routingKey:   routingKey,
	}
}

func startChunkSender(id int, channel <-chan ChunkItem, broker amqp.MessageBroker, exchange string, chunkMaxSize uint8, routingKey string) {
	s := newChunkSender(id, channel, broker, exchange, chunkMaxSize, routingKey)
	for {
		item := <-channel
		s.updateChunk(item.Msg, item.Msg == nil)
	}
}

func (s *ChunkSender) updateChunk(data any, eof bool) {
	if !eof {
		s.chunk = append(s.chunk, data)
	}

	s.sendChunk(eof)
}

// sendChunk sends a chunk of data to the broker if the chunk is full or the eof flag is true
// In case eof is true, it sends an EOF message to the broker
// if a chunk was sent restarts count
func (s *ChunkSender) sendChunk(eof bool) {
	messageId := matchMessageId(s.id)
	if len(s.chunk) > 0 && (len(s.chunk) >= int(s.maxChunkSize) || eof) {
		bytes, err := toBytes(messageId, s.chunk)
		if err != nil {
			logs.Logger.Errorf("Error converting chunk to bytes: %s", err.Error())
		}

		err = s.broker.Publish(s.exchange, s.routingKey, bytes, map[string]any{amqp.MessageIdHeader: uint8(messageId)})
		if err != nil {
			logs.Logger.Errorf("Error publishing chunk: %s", err.Error())
		}

		s.chunk = make([]any, 0, s.maxChunkSize)
	}
	if eof {
		err := s.broker.Publish(s.exchange, s.routingKey, amqp.EmptyEof, map[string]any{amqp.MessageIdHeader: uint8(message.EofMsg)})
		if err != nil {
			logs.Logger.Errorf("Error publishing EOF: %s", err.Error())
		}
		logs.Logger.Infof("Sent eof with key %v", s.routingKey)
		s.chunk = make([]any, 0, s.maxChunkSize)
	}
}

func toBytes(msgId message.ID, chunk []any) ([]byte, error) {
	if msgId == message.ReviewIdMsg {
		reviews := make([]message.DataCSVReviews, 0, len(chunk))
		for _, v := range chunk {
			reviews = append(reviews, v.(message.DataCSVReviews))
		}
		return message.ReviewsFromClientReviews(reviews)
	}
	games := make([]message.DataCSVGames, 0, len(chunk))
	for _, v := range chunk {
		games = append(games, v.(message.DataCSVGames))
	}
	return message.GamesFromClientGames(games)
}
