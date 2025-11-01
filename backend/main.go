// go\
package main

import (
	_ "bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	_ "github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func makeTurn(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(400)
		return
	}
	playerUUID := uuid.MustParse(token.Value)
	_, index := getPlayerByUUID(playerUUID)

	if gameInfo.GameInfo.Turn != index {
		w.WriteHeader(400)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read request")
		w.WriteHeader(500)
		return
	}

	log.Println(string(bytes))
	path := PathDecodePathFromJson(bytes)
	word, ok := ValidPath(wordset, &gameInfo, path)
	if !ok {
		log.Println("invalid path")

		response := "{\"isWord\": false, \"score\": 0}"
		w.Write([]byte(response))
		return
	}

	score, gems := ScorePath(&gameInfo, path)
	response := fmt.Sprintf("{\"isWord\": true, \"score\": %d}", score)
	_, err = w.Write([]byte(response))
	if err != nil {
		w.WriteHeader(500)
		log.Println("failed to write response")
		return
	}

	gameInfo.PlayerInfo.Players[gameInfo.GameInfo.Turn].PlayerScore += score
	gameInfo.PlayerInfo.Players[gameInfo.GameInfo.Turn].PlayerGems += gems
	wordPlayed := WordPlayed{Word: word, Score: score, PlayerName: gameInfo.PlayerInfo.Players[gameInfo.GameInfo.Turn].PlayerName}

	gameInfo.GameInfo.WordsPlayed = append(gameInfo.GameInfo.WordsPlayed, wordPlayed)

	gameInfo.GameInfo.Turn++
	if gameInfo.GameInfo.Turn >= len(gameInfo.PlayerInfo.Players) {
		gameInfo.GameInfo.Turn = 0
		gameInfo.GameInfo.GameTurn += 1
	}

	GameRandomisePath(&gameInfo, path, gameInfo.GameInfo.GameTurn > 0, gameInfo.GameInfo.GameTurn > 0)
	SendNewGameInfo()
}

func shuffle(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(400)
		return
	}
	playerUUID := uuid.MustParse(token.Value)
	player, index := getPlayerByUUID(playerUUID)

	if gameInfo.GameInfo.Turn != index {
		w.Write([]byte("{\"message\": Not Your Turn}"))
		return
	}
	if player.PlayerGems < 1 {
		w.Write([]byte("{\"message\": Not Enough Gems}"))
		return
	}

	gameInfo.PlayerInfo.Players[index].PlayerGems -= 1

	GameScrambleBoard(&gameInfo)
	SendNewGameInfo()
}

func swap(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(400)
		return
	}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read request")
		w.WriteHeader(500)
		return
	}

	var info struct {
		Pos    Position `json:"position"`
		Letter Letter   `json:"letter"`
	}

	err = json.Unmarshal(bytes, &info)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	playerUUID := uuid.MustParse(token.Value)
	player, index := getPlayerByUUID(playerUUID)

	if gameInfo.GameInfo.Turn != index {
		w.Write([]byte("{\"message\": Not Your Turn}"))
		return
	}
	if player.PlayerGems < 3 {
		w.Write([]byte("{\"message\": Not Enough Gems}"))
		return
	}

	gameInfo.PlayerInfo.Players[index].PlayerGems -= 3
	GameSwapTile(&gameInfo, info.Pos, info.Letter)
	SendNewGameInfo()
}

func SendNewGameInfo() {
	for index, playerData := range gameInfo.PlayerInfo.Players {
		conn := playerData.playerConn
		if conn == nil {
			continue
		}
		var sendInfo struct {
			YourTurn int       `json:"yourTurn"`
			GameData *GameInfo `json:"gameInfo"`
		}

		sendInfo.YourTurn = index
		sendInfo.GameData = &gameInfo

		err := conn.WriteJSON(sendInfo)
		if err != nil {
			conn.Close()
			conn = nil
			log.Println(err)
			continue
		}
	}
}

func getBoard(w http.ResponseWriter, r *http.Request) {
	log.Println("board get")
	type BoardJson struct {
		Board [6][6]Letter `json:"board"`
	}
	bytes, err := json.Marshal(BoardJson{Board: gameInfo.BoardInfo.Board})
	if err != nil {
		log.Println("error marshalling board to json")
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Println("failed to write response")
		w.WriteHeader(500)
		return
	}
}

type Turn struct {
	PlayerId int        `json:"id"`
	Path     []Position `json:"path"`
}

var wordset WordSet

func initWordset() {
	wordlist := GetWordList("wordlist.txt")
	wordset = make(map[string]struct{})

	for _, word := range wordlist {
		wordset[word] = struct{}{}
	}
}
func reset(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(400)
		return
	}
	playerUUID := uuid.MustParse(token.Value)
	_, index := getPlayerByUUID(playerUUID)

	if index != 0 {
		w.Write([]byte("{\"message\": Only Host Can Reset}"))
		return
	}

	initGame(&gameInfo)
	SendNewGameInfo()
}
func kick(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(400)
		return
	}
	playerUUID := uuid.MustParse(token.Value)
	_, index := getPlayerByUUID(playerUUID)

	if index != 0 {
		w.Write([]byte("{\"message\": Only Host Can Reset}"))
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read request")
		w.WriteHeader(500)
		return
	}

	var info struct {
		Index int `json:"index"`
	}

	err = json.Unmarshal(bytes, &info)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	gameInfo.PlayerInfo.Players[info.Index].playerConn.Close()

	s := gameInfo.PlayerInfo.Players
	gameInfo.PlayerInfo.Players = append(s[:info.Index], s[info.Index+1:]...)
	if gameInfo.GameInfo.Turn > len(gameInfo.PlayerInfo.Players) {
		gameInfo.GameInfo.Turn %= len(gameInfo.PlayerInfo.Players)
	}
	SendNewGameInfo()
}

func main() {
	initWordset()
	initGame(&gameInfo)
	gameInfo.GameInfo.WordsPlayed = []WordPlayed{}

	server := &http.Server{
		Addr: ":8080",
	}
	go func() {

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		<-signalChan
		if err := server.Close(); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}
	}()

	// Serve static files from frontend build
	fs := os.DirFS("../frontend/dist")
	http.Handle("/", http.FileServerFS(fs))

	http.HandleFunc("/session", session)
	http.HandleFunc("/join", join)
	http.HandleFunc("/setName", setName)
	http.HandleFunc("/board", getBoard)
	http.HandleFunc("/turn", makeTurn)
	http.HandleFunc("/shuffle", shuffle)
	http.HandleFunc("/swap", swap)
	http.HandleFunc("/reset", reset)
	http.HandleFunc("/kick", kick)

	log.Printf("listening on port %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
