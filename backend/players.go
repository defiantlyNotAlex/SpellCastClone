package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func session(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session")
	if err != nil {
		playerUUID := uuid.New()
		token := http.Cookie{
			Name:     "session",
			Value:    playerUUID.String(),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, &token)
	}
}

func join(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(400)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	log.Println("Websocket Opened")

	log.Println(token.Value)
	playerUUID := uuid.MustParse(token.Value)

	player, _ := getPlayerByUUID(playerUUID)

	if player != nil {
		log.Println("reconnect")

		player.playerConn = conn
	} else {
		log.Println("new player")
		player := PlayerData{
			playerId:    playerUUID,
			playerConn:  conn,
			PlayerName:  "no name",
			PlayerScore: 0,
			PlayerGems:  3,
		}

		gameInfo.PlayerInfo.Players = append(gameInfo.PlayerInfo.Players, player)
	}

	SendNewGameInfo()
}

func setName(w http.ResponseWriter, r *http.Request) {
	tokens := r.CookiesNamed("session")
	if len(tokens) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := tokens[0]
	if err := token.Valid(); err != nil {
		w.WriteHeader(500)
		log.Println(err)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Panicln(err)
		w.WriteHeader(500)
		return
	}

	var name struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(bytes, &name)
	if err != nil {
		log.Panicln(err)
		w.WriteHeader(500)
		return
	}

	playerUUID, err := uuid.Parse(token.Value)
	if err != nil {
		w.WriteHeader(500)
		log.Println(err)
		return
	}

	player, index := getPlayerByUUID(playerUUID)

	player.PlayerName = name.Name
	gameInfo.PlayerInfo.Players[index].PlayerName = name.Name
	SendNewGameInfo()
}

func getPlayerByUUID(playerId uuid.UUID) (*PlayerData, int) {
	for i, player := range gameInfo.PlayerInfo.Players {
		if player.playerId == playerId {
			return &gameInfo.PlayerInfo.Players[i], i
		}
	}
	return nil, 0
}
