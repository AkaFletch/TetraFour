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
	Score            int32
	Board            [11]byte
	Piece            [11][7]int8
	NextPiece        Tetra
	CurrentPieceType Tetra
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
	log.Info().Msg("New Tweet")
	log.Debug().Msg(tweet.Text)
	if tweet.User.IDStr == config.Twitter.UserId {
		log.Debug().Msgf("Ignoring tweet not from %s", config.Twitter.UserId)
		return
	}
	state := TextToBoard(tweet.Text)
	state.printBoard()
	state.FindCurrentPiece(tweet)
	score := state.ScoreBoard()
	log.Info().Msgf("Current Scored %d", score)
	log.Debug().Msgf("Next piece: %s", state.NextPiece.String())
}

func ParseNextPiece(boardText string) Tetra {
	board := strings.Split(boardText, "\n")
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

func TextToBoard(text string) GameState {
	state := GameState{}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		val := ParseLine(line)
		state.Board[i] = val
	}
	state.NextPiece = ParseNextPiece(text)
	return state
}

func (board GameState) FindCurrentPiece(tweet *twitter.Tweet) {
	// go back until the next piece changes, or a + is on the board
	lastTweetId := tweet.InReplyToStatusID
	var pastTweet *twitter.Tweet
	var nextPieceType Tetra
	foundCurrent := false
	for !foundCurrent {
		pastTweet = GetTweetBefore(lastTweetId)
		if strings.Contains(pastTweet.Text, "+") {
			earlierTweet := GetTweetBefore(pastTweet.InReplyToStatusID)
			earlierBoard := TextToBoard(earlierTweet.Text)
			nextPieceType = earlierBoard.NextPiece
			foundCurrent = true
		}
		pastBoard := TextToBoard(pastTweet.Text)
		if pastBoard.NextPiece != board.NextPiece {
			nextPieceType = pastBoard.NextPiece
			foundCurrent = true
		}
		lastTweetId = pastTweet.InReplyToStatusID
	}
	log.Info().Msgf("Current Piece: %s", nextPieceType.String())
}
