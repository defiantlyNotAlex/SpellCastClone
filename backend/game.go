package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type PathJson struct {
	Data []Position `json:"path"`
}

func Abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}
func PositionEqual(p1 Position, p2 Position) bool {
	return p1.X == p2.X && p1.Y == p2.Y
}
func PositionAdjacent(p1 Position, p2 Position) bool {
	return Abs(p1.X-p2.X) <= 1 && Abs(p1.Y-p2.Y) <= 1
}
func PositionRand() Position {
	return Position{X: rand.Intn(6), Y: rand.Intn(6)}
}

type Letter int

var gameInfo GameInfo

type PlayerData struct {
	PlayerName  string `json:"name"`
	PlayerScore int    `json:"score"`
	PlayerGems  int    `json:"gems"`
	playerId    uuid.UUID
	playerConn  *websocket.Conn
}
type GameInfo struct {
	BoardInfo struct {
		Board            [6][6]Letter `json:"board"`
		DoubleLetter     Position     `json:"doubleLetter"`
		DoubleWord       Position     `json:"doubleWord"`
		DoubleLetterMult int          `json:"doubleLetterMult"`
	} `json:"boardInfo"`
	PlayerInfo struct {
		Players []PlayerData `json:"players"`
	} `json:"playerInfo"`
	GameInfo struct {
		Turn        int          `json:"turn"`
		GameTurn    int          `json:"gameTurn"`
		GameEnded   bool         `json:"gameEnded"`
		WordsPlayed []WordPlayed `json:"wordsPlayed"`
	} `json:"gameInfo"`
}

type WordPlayed struct {
	Word       string `json:"word"`
	Score      int    `json:"score"`
	PlayerName string `json:"playerName"`
}

type WordSet map[string]struct{}

func PathDecodePathFromJson(b []byte) []Position {
	var pathJson PathJson
	err := json.Unmarshal(b, &pathJson)
	if err != nil {
		log.Println(err)
		return nil
	}

	return pathJson.Data
}

var alphabet = [...]rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
var vowels = [...]rune{'a', 'e', 'i', 'o', 'u'}
var consonents = [...]rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z'}
var letterScores = [...]int{1, 4, 5, 3, 1, 5, 3, 4, 1, 7, 6, 3, 4, 2, 1, 4, 8, 2, 2, 2, 4, 5, 5, 7, 4, 8}

func RandomLetter() Letter {
	n := rand.Intn(26)
	for rand.Intn(letterScores[n]) != 0 {
		n = rand.Intn(26)
	}
	return Letter(n)
}

func GameRandomiseBoard(game *GameInfo, tripleLetter bool, doubleWord bool) {
	for y := range 6 {
		for x := range 6 {
			game.BoardInfo.Board[y][x] = RandomLetter()
		}
	}
	if tripleLetter {
		game.BoardInfo.DoubleLetterMult = 3
	} else {
		game.BoardInfo.DoubleLetterMult = 2
	}
	game.BoardInfo.DoubleLetter = PositionRand()
	if doubleWord {
		game.BoardInfo.DoubleWord = PositionRand()
	} else {
		game.BoardInfo.DoubleWord = Position{X: -1, Y: -1}
	}
}

func GameRandomisePath(game *GameInfo, path []Position, tripleLetter bool, doubleWord bool) {

	if doubleWord {
		if slices.Contains(path, game.BoardInfo.DoubleWord) {
			game.BoardInfo.DoubleWord = PositionRand()
			for PositionEqual(game.BoardInfo.DoubleWord, game.BoardInfo.DoubleLetter) {
				game.BoardInfo.DoubleWord = PositionRand()
			}
		}
		if PositionEqual(game.BoardInfo.DoubleWord, Position{-1, -1}) {
			game.BoardInfo.DoubleWord = PositionRand()
		}
	}
	if slices.Contains(path, game.BoardInfo.DoubleLetter) {
		game.BoardInfo.DoubleLetter = PositionRand()
		for PositionEqual(game.BoardInfo.DoubleWord, game.BoardInfo.DoubleLetter) {
			game.BoardInfo.DoubleLetter = PositionRand()
		}
		if !tripleLetter || rand.Int()%2 == 0 {
			game.BoardInfo.DoubleLetterMult = 2
		} else {
			game.BoardInfo.DoubleLetterMult = 3
		}
	}

	for _, p := range path {
		game.BoardInfo.Board[p.Y][p.X] = RandomLetter()
	}
}

func GameTurn(game *GameInfo, path []Position) bool {
	if _, ok := ValidPath(wordset, game, path); !ok {
		return false
	}
	score := ScorePath(game, path)
	game.PlayerInfo.Players[game.GameInfo.Turn].PlayerScore += score

	game.GameInfo.Turn++
	if game.GameInfo.Turn > len(game.PlayerInfo.Players) {
		game.GameInfo.Turn = 0
		game.GameInfo.GameTurn += 1
	}

	GameRandomisePath(game, path, game.GameInfo.GameTurn > 0, game.GameInfo.GameTurn > 0)

	return true
}

func GameScrambleBoard(game *GameInfo) {
	for y := range 6 {
		for x := range 6 {
			i := x + y*6
			if i == 35 {
				return
			}
			j := rand.Intn(36-i) + i
			rx := j % 6
			ry := j / 6

			if game.BoardInfo.DoubleLetter.X == rx && game.BoardInfo.DoubleLetter.Y == ry {
				game.BoardInfo.DoubleLetter.X = x
				game.BoardInfo.DoubleLetter.Y = y
			} else if game.BoardInfo.DoubleLetter.X == x && game.BoardInfo.DoubleLetter.Y == y {
				game.BoardInfo.DoubleLetter.X = rx
				game.BoardInfo.DoubleLetter.Y = ry
			}

			if game.BoardInfo.DoubleWord.X == rx && game.BoardInfo.DoubleWord.Y == ry {
				game.BoardInfo.DoubleWord.X = x
				game.BoardInfo.DoubleWord.Y = y
			} else if game.BoardInfo.DoubleWord.X == x && game.BoardInfo.DoubleWord.Y == y {
				game.BoardInfo.DoubleWord.X = rx
				game.BoardInfo.DoubleWord.Y = ry
			}

			temp := game.BoardInfo.Board[y][x]

			game.BoardInfo.Board[y][x] = game.BoardInfo.Board[ry][rx]
			game.BoardInfo.Board[ry][rx] = temp
		}
	}
}

func GameSwapTile(game *GameInfo, p Position, l Letter) {
	game.BoardInfo.Board[p.Y][p.X] = l
}

func ValidPath(wordset WordSet, game *GameInfo, path []Position) (string, bool) {

	var wordBuilder strings.Builder
	used := [6][6]bool{
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
	}

	if len(path) == 0 || len(path) > 6*6 {
		return "", false
	}

	last := path[0]
	for _, p := range path {

		// already used the tile
		if used[p.Y][p.X] {
			return "", false
		}
		// tile adjacent to the last tile in the path
		if !PositionAdjacent(last, p) {
			return "", false
		}
		used[p.Y][p.X] = true
		wordBuilder.WriteRune(alphabet[game.BoardInfo.Board[p.Y][p.X]])

		last = p
	}
	word := wordBuilder.String()
	_, ok := wordset[word]

	return word, ok
}

func ScorePath(game *GameInfo, path []Position) int {

	doubleScore := false

	score := 0

	for _, p := range path {
		letter := game.BoardInfo.Board[p.Y][p.X]
		letterScore := letterScores[letter]
		if PositionEqual(game.BoardInfo.DoubleLetter, p) {
			letterScore *= game.BoardInfo.DoubleLetterMult
		}
		if PositionEqual(game.BoardInfo.DoubleWord, p) {
			doubleScore = true
		}
		score += letterScore
	}

	if doubleScore {
		score *= 2
	}
	if len(path) >= 6 {
		score += 10
	}

	return score
}
