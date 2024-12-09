package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"tp1/pkg/message"
)

func TestGameNameToBytes(t *testing.T) {
	gn := message.GameName{
		GameId:   12345,
		GameName: "Test Game",
	}

	data, err := gn.ToBytes()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	decodedGameName, err := message.GameNameFromBytes(data)
	assert.NoError(t, err)
	assert.Equal(t, gn.GameId, decodedGameName.GameId)
	assert.Equal(t, gn.GameName, decodedGameName.GameName)
}

func TestGameNamesToBytesFromBytes(t *testing.T) {
	gn := message.GameNames{
		{GameId: 12345, GameName: "Test Game 1"},
		{GameId: 67890, GameName: "Test Game 2"},
	}

	data, err := gn.ToBytes()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	decodedGameNames, err := message.GameNamesFromBytes(data)
	assert.NoError(t, err)
	assert.Equal(t, len(gn), len(decodedGameNames))

	for i, game := range gn {
		assert.Equal(t, game.GameId, decodedGameNames[i].GameId)
		assert.Equal(t, game.GameName, decodedGameNames[i].GameName)
	}
}

func TestToStringAux(t *testing.T) {
	gn := message.GameNames{
		{GameId: 12345, GameName: "Test Game 1"},
		{GameId: 67890, GameName: "Test Game 2"},
	}

	expected := "Juego: [Test Game 1], Id: [12345]\nJuego: [Test Game 2], Id: [67890]\n"
	actual := gn.ToStringAux()

	assert.Equal(t, expected, actual)
}

func TestGameNamesFromBytesError(t *testing.T) {
	invalidData := []byte{0x01, 0x02, 0x03}

	_, err := message.GameNamesFromBytes(invalidData)
	assert.Error(t, err)
}
