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

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type PlayerList struct {
	playerId         []int
	playerNames      []string
	playerConnection []*websocket.Conn
}

var playerList PlayerList
var id int = 0

func join(w http.ResponseWriter, r *http.Request) {
	log.Println("Websocket Opened")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	_ = conn

	err = conn.WriteJSON(gameInfo)
	if err != nil {
		log.Println(err)
		return
	}

	// TODO: add a player id system for session tokens
	// TODO: add a way of players inputing a name
	// TODO: add a way for players to change their name

	playerList.playerId = append(playerList.playerId, id)
	id++
	playerList.playerNames = append(playerList.playerNames, "frank")
	playerList.playerConnection = append(playerList.playerConnection, conn)

	gameInfo.PlayerInfo.PlayerNames = playerList.playerNames
	gameInfo.PlayerInfo.PlayerScores = append(gameInfo.PlayerInfo.PlayerScores, 0)
}

func readLoop(c *websocket.Conn) {
	for {
		if messageType, r, err := c.NextReader(); err != nil || messageType != websocket.TextMessage {
			c.Close()
			break
		} else {
			bytes, err := io.ReadAll(r)
			if err != nil {
				log.Panicln(err)
				continue
			}
			log.Panicln(string(bytes))
		}
	}
}

func makeTurn(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read request")
		w.WriteHeader(500)
		return
	}

	log.Println(string(bytes))
	path := PathDecodePathFromJson(bytes)
	if !ValidPath(wordset, &gameInfo, path) {
		log.Println("invalid path")

		response := "{\"is_word\": false, \"score\": 0}"
		w.Write([]byte(response))
		return
	}

	score := ScorePath(&gameInfo, path)
	response := fmt.Sprintf("{\"is_word\": true, \"score\": %d}", score)
	_, err = w.Write([]byte(response))
	if err != nil {
		w.WriteHeader(500)
		log.Println("failed to write response")
		return
	}

	gameInfo.PlayerInfo.PlayerScores[gameInfo.GameInfo.Turn] += score

	gameInfo.GameInfo.Turn++
	if gameInfo.GameInfo.Turn >= len(gameInfo.PlayerInfo.PlayerNames) {
		gameInfo.GameInfo.Turn = 0
		gameInfo.GameInfo.GameTurn += 1
	}

	GameRandomisePath(&gameInfo, path, gameInfo.GameInfo.GameTurn > 0, gameInfo.GameInfo.GameTurn > 0)
	SendNewGameInfo()
}

func SendNewGameInfo() {
	for _, conn := range playerList.playerConnection {
		if conn == nil {
			continue
		}
		err := conn.WriteJSON(gameInfo)
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

func init_wordset() {
	wordlist := GetWordList("wordlist.txt")
	wordset = make(map[string]struct{})

	for _, word := range wordlist {
		wordset[word] = struct{}{}
	}
}

func main() {
	init_wordset()
	GameRandomiseBoard(&gameInfo, false, false)

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
	fs := http.FileServer(http.Dir("../frontend/dist"))
	http.Handle("/", fs)

	http.HandleFunc("/join", join)
	http.HandleFunc("/board", getBoard)
	http.HandleFunc("/turn", makeTurn)

	log.Printf("listening on port %s", server.Addr)
	log.Fatal(server.ListenAndServe())

}
