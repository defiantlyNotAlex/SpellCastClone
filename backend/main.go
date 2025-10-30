// go\
package main

import (
	_ "bufio"
	"encoding/json"
	"fmt"
	"io"
	_ "io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

func handler(w http.ResponseWriter, r *http.Request) {
	bytes, err := os.ReadFile("../front/test.html")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	n, err := w.Write(bytes)

	_ = n

	if err != nil {
		w.WriteHeader(500)
		return
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func join(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	_ = conn

}

func readLoop(c *websocket.Conn) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			c.Close()
			break
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
	if !ValidPath(wordset, &game, path) {
		log.Println("invalid path")

		response := "{\"is_word\": false, \"score\": 0}"
		w.Write([]byte(response))
		return
	}

	score := ScorePath(&game, path)
	response := fmt.Sprintf("{\"is_word\": true, \"score\": %d}", score)
	_, err = w.Write([]byte(response))
	if err != nil {
		w.WriteHeader(500)
		log.Println("failed to write response")
		return
	}

}

func getBoard(w http.ResponseWriter, r *http.Request) {
	log.Println("board get")
	type BoardJson struct {
		Board Board `json:"board"`
	}
	bytes, err := json.Marshal(BoardJson{Board: game.board})
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
	PlayerId int  `json:"id"`
	Path     Path `json:"path"`
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
	GameRandomiseBoard(&game)

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
