package gateway

import (
	"fmt"
	"github.com/pierrec/xxHash/xxHash32"

	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type ChunkSender struct {
	id           int //0 for games, 1 for reviews
	channel      <-chan ChunkItem
	broker       amqp.MessageBroker
	dst          []amqp.Destination
	maxChunkSize uint8
	chunks       map[string][]any
}

type ChunkItem struct {
	Msg      any //DataCSVGames or DataCSVReviews
	ClientId string
}

func newChunkSender(id int, channel <-chan ChunkItem, broker amqp.MessageBroker, dst []amqp.Destination, chunkMaxSize uint8) *ChunkSender {
	return &ChunkSender{
		id:           id,
		channel:      channel,
		broker:       broker,
		dst:          dst,
		chunks:       make(map[string][]any),
		maxChunkSize: chunkMaxSize,
	}
}

func startChunkSender(id int, channel <-chan ChunkItem, broker amqp.MessageBroker, dst []amqp.Destination, chunkMaxSize uint8) {
	s := newChunkSender(id, channel, broker, dst, chunkMaxSize)
	for {
		item := <-channel
		s.updateChunk(item, item.Msg == nil)
	}
}

func (s *ChunkSender) updateChunk(item ChunkItem, eof bool) {
	clientId := item.ClientId
	if !eof {
		if _, ok := s.chunks[clientId]; !ok {
			s.chunks[clientId] = make([]any, 0, s.maxChunkSize)
		}
		s.chunks[clientId] = append(s.chunks[clientId], item)
	}

	s.sendChunk(eof, clientId)
}

// sendChunk sends a chunks of data to the broker if the chunks is full or the eof flag is true
// In case eof is true, it sends an EOF message to the broker
// if a chunks was sent restarts count
func (s *ChunkSender) sendChunk(eof bool, clientId string) {
	messageId := matchMessageId(s.id)
	chunk := s.chunks[clientId]

	if len(chunk) >= int(s.maxChunkSize) || eof {
		bytes, err := toBytes(messageId, chunk)
		if err != nil {
			logs.Logger.Errorf("Error converting chunks to bytes: %s", err.Error())
		}

		headers := amqp.Header{MessageId: messageId, ClientId: clientId}

		if err = s.publish(bytes, headers); err != nil {
			logs.Logger.Errorf("Error publishing chunks: %s", err.Error())
		}

		s.chunks[clientId] = make([]any, 0, s.maxChunkSize)
	}

	if eof {
		headers := amqp.Header{MessageId: message.EofMsg, ClientId: clientId}

		if err := s.publish(amqp.EmptyEof, headers); err != nil {
			logs.Logger.Errorf("Error publishing EOF: %s", err.Error())
		}

		delete(s.chunks, clientId)
	}
}

func (s *ChunkSender) publish(msg []byte, headers amqp.Header) error {
	for _, dst := range s.dst {
		key := shardSequenceId(headers.SequenceId, dst.Key, dst.Consumers)
		if err := s.broker.Publish(dst.Exchange, key, msg, headers); err != nil {
			return err
		}
	}
	return nil
}

func toBytes(msgId message.ID, chunk []any) ([]byte, error) {
	if msgId == message.ReviewIdMsg {
		reviews := make([]message.DataCSVReviews, 0, len(chunk))
		for _, v := range chunk {
			reviews = append(reviews, v.(ChunkItem).Msg.(message.DataCSVReviews))
		}
		return message.ReviewsFromClientReviews(reviews)
	}
	games := make([]message.DataCSVGames, 0, len(chunk))
	for _, v := range chunk {
		games = append(games, v.(ChunkItem).Msg.(message.DataCSVGames))
	}
	return message.GamesFromClientGames(games)
}

func shardSequenceId(id string, key string, consumers uint8) string {
	if consumers == 0 {
		return key
	}
	return fmt.Sprintf(key, xxHash32.Checksum([]byte(id), 0)%uint32(consumers))
}
