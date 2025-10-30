package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"slices"
	"strings"
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

type Path []Position
type Board [6][6]rune

var game Game

type Game struct {
	board Board
	dl    Position
	dw    Position

	letter_multiplier_3x bool

	turn         int
	playerCount  int
	playerScores []int
}

type WordSet map[string]struct{}

func PathDecodePathFromJson(b []byte) Path {
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

func RandomLetter() rune {
	n := rand.Intn(26)
	for rand.Intn(letterScores[n]) != 0 {
		n = rand.Intn(26)
	}
	return alphabet[n]
}

func GameRandomiseBoard(game *Game) {
	for y := range 6 {
		for x := range 6 {
			game.board[y][x] = RandomLetter()
		}
	}
}

func GameTurn(game *Game, path Path) bool {
	if !ValidPath(wordset, game, path) {
		return false
	}
	score := ScorePath(game, path)
	game.playerScores[game.playerCount] = score

	if slices.Contains(path, game.dw) {
		game.dw = PositionRand()
		for PositionEqual(game.dw, game.dl) {
			game.dw = PositionRand()
		}
	}
	if slices.Contains(path, game.dl) {
		game.dl = PositionRand()
		for PositionEqual(game.dw, game.dl) {
			game.dl = PositionRand()
		}
	}

	for _, p := range path {
		game.board[p.Y][p.X] = RandomLetter()
	}

	game.turn++
	game.turn %= game.playerCount

	return true
}

func GameScrambleBoard(game *Game) {

	var temp rune

	for y := range 6 {
		for x := range 6 {
			i := x + y*6
			if i == 35 {
				return
			}
			j := rand.Intn(36-i) + i
			rx := j % 6
			ry := j / 6

			if game.dl.X == rx && game.dl.Y == ry {
				game.dl.X = x
				game.dl.Y = y
			}
			if game.dl.X == x && game.dl.Y == y {
				game.dl.X = rx
				game.dl.Y = ry
			}

			if game.dw.X == rx && game.dw.Y == ry {
				game.dw.X = x
				game.dw.Y = y
			}
			if game.dw.X == x && game.dw.Y == y {
				game.dw.X = rx
				game.dw.Y = ry
			}

			temp = game.board[y][x]

			game.board[y][x] = game.board[ry][rx]
			game.board[ry][rx] = temp
		}
	}
}

func SwapTile(board *Board, p Position, r rune) {
	(*board)[p.Y][p.X] = r
}

func ValidPath(wordset WordSet, game *Game, path Path) bool {

	var word strings.Builder
	used := [6][6]bool{
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
		{false, false, false, false, false, false},
	}

	if len(path) == 0 || len(path) > 6*6 {
		return false
	}

	last := path[0]
	for _, p := range path {

		// already used the tile
		if used[p.Y][p.X] {
			return false
		}
		// tile adjacent to the last tile in the path
		if !PositionAdjacent(last, p) {
			return false
		}
		used[p.Y][p.X] = true
		word.WriteRune(game.board[p.Y][p.X])

		last = p
	}

	_, ok := wordset[word.String()]

	return ok
}

func ScorePath(game *Game, path Path) int {

	doubleScore := false

	score := 0

	for _, p := range path {
		char := game.board[p.Y][p.X]
		letterScore := letterScores[int(char-'a')]
		if PositionEqual(game.dl, p) {
			if game.letter_multiplier_3x {
				letterScore *= 3
			} else {
				letterScore *= 2
			}
		}
		if PositionEqual(game.dw, p) {
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
