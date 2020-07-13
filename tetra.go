package main

import (
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/rs/zerolog/log"
)

var EmptyCharacter = '◽'

type Tetra int

const (
	N Tetra = iota
	Line
	Block
	L
	T
	J
	Z
)

var NextTetra = map[byte]Tetra{
	byte(0b10110100): N,
	byte(0b10101010): Line,
	byte(0b11110000): Block,
	byte(0b10101100): L,
	byte(0b10111000): T,
	byte(0b11101000): J,
	byte(0b01111000): Z,
}

type GameState struct {
	Score int32
	Board [11]byte
	Piece [11][7]int8
}

func (t Tetra) String() string {
	return [...]string{
		"N",
		"Line",
		"Block",
		"L",
		"T",
		"J",
		"Z",
	}[t]
}

func (state *GameState) printBoard() {
	for _, line := range state.Board {
		log.Debug().Msgf("%0.8b", line)
	}
}

func ParseTweet(tweet *twitter.Tweet) {
	log.Debug().Msg("New Tweet")
	lines := strings.Split(tweet.Text, "\n")
	state := GameState{}
	for i, line := range lines {
		val := ParseLine(line)
		state.Board[i] = val
	}
	state.printBoard()
	ParseNextPiece(lines)
	score := state.ScoreBoard()
	log.Info().Msgf("Current Scored %d", score)
}

func ParseNextPiece(board []string) Tetra {
	madeByte := byte(0)
	for i, line := range board {
		if 2 > i || i > 5 {
			continue
		}
		pos := 0
		// Not using the range postion since unicode chars are multiple bytes
		for _, char := range line {
			if pos < 8 || char == EmptyCharacter {
				pos++
				continue
			}
			madeByte |= byte(128) >> ((i-2)*2 + pos - 8)
			pos++
		}
	}
	log.Debug().Msgf("Next piece: %s", NextTetra[madeByte].String())
	return NextTetra[madeByte]
}

func ParseLine(line string) byte {
	returnByte := byte(0)
	pos := 0
	for _, char := range line {
		if pos > 6 {
			continue
		}
		occupiedSpace := char != EmptyCharacter
		shift := BoolToInt(occupiedSpace)
		returnByte |= byte(byte(128*shift) >> pos)
		pos++
	}
	return returnByte
}

func (board GameState) ScoreSideways(y, dir int) int8 {
	health := int8(0)
	for i := 0; i < 8; i++ {
		if i+dir > 7 || i+dir < 0 {
			continue
		}
		checkMask := byte(1) << uint(i)
		testMask := byte(1) << uint(i+dir)
		if board.Board[y]&checkMask > 1 && board.Board[y]&testMask > 1 {
			health += 1
		}
	}
	return health
}

func (board GameState) ScoreBoard() int8 {
	var left, right int8
	for i := 10; i > -1; i-- {
		left += board.ScoreSideways(i, -1)
		right += board.ScoreSideways(i, 1)
	}
	return left + right
}
