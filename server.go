package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

type LobbyJson struct {
	Ready string `json:"ready"`
	Name  string `json:"name"`
	Job   string `json:"job"`
	ID    int    `json:"ID"`
}

type StartGameJson struct {
	Server string `json:"server"`
	Name   string `json:"name"`
}

type Player struct {
	ID           int
	IsConnection bool
	Websocket    *melody.Session
	Ready        string
	_LobbyJson   *LobbyJson
}

func main() {
	r := gin.Default()
	m := melody.New()
	countID := 0

	connections := map[int]*Player{}

	m.HandleConnect(func(session *melody.Session) {
		fmt.Println("接続しました。")
		countID += 1
		connections[countID] = &Player{ID: countID, IsConnection: true, Websocket: session, Ready: "no"}
		message := LobbyJson{ID: countID, Name: "my"}
		messageJson, err := json.Marshal(message)

		if err != nil {
			panic(err)
		}

		connections[countID].Websocket.Set("Content-Type", "application/json")
		connections[countID].Websocket.Write([]byte(messageJson))

	})

	m.HandleDisconnect(func(session *melody.Session) {
		fmt.Println("切断されました。")
	})

	m.HandleMessage(func(session *melody.Session, message []byte) {

		var lobby_json LobbyJson
		json.Unmarshal(message, &lobby_json)
		fmt.Println(lobby_json)
		connections[lobby_json.ID]._LobbyJson = &lobby_json
		if lobby_json.Name == "my" {
			lobby_json.Name = "opponent"
		} else {
			lobby_json.Name = "my"
		}

		if lobby_json.Ready == "ready" {
			connections[lobby_json.ID].Ready = "ready"
		}

		opponentJSON, err := json.Marshal(lobby_json)
		if err != nil {
			panic(err)
		}

		if lobby_json.ID%2 == 1 {

			if connections[lobby_json.ID+1] != nil {
				connections[lobby_json.ID+1].Websocket.Write(opponentJSON)
			}

		} else {

			if connections[lobby_json.ID-1] != nil {
				connections[lobby_json.ID-1].Websocket.Write(opponentJSON)
			}
		}
		/*
			check_json, err := json.Marshal(connections[lobby_json.ID]._LobbyJson)

			if err != nil {
				panic(err)
			}
			connections[lobby_json.ID].Websocket = session
			connections[lobby_json.ID].Websocket.Write(check_json)
		*/

		//readygoの処理
		if connections[1] != nil && connections[2] != nil {
			if connections[1].Ready == "ready" && connections[2].Ready == "ready" {

				fmt.Println("both ready")

				connections[1].Ready, connections[2].Ready = "no", "no"
				start_message := StartGameJson{Server: "readygo", Name: "server"}
				start_json, err := json.Marshal(start_message)

				if err != nil {
					panic(err)
				}
				connections[1].Websocket.Set("Content-Type", "application/json")
				connections[2].Websocket.Set("Content-Type", "application/json")

				time.Sleep(3500)

				connections[1].Websocket.Write(start_json)
				connections[2].Websocket.Write(start_json)

			}
		}
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	r.Run(":8088")

}
