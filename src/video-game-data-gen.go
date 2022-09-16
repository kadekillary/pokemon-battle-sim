package main

import (
	"context"
	"encoding/json"
	"log"
    "fmt"
    "strconv"
	//"math/rand"
	//"time"

	//"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
)

//type Player struct {
//	Id   int    `json:"id"`
//	Name string `json:"name"`
//}

//type Score struct {
//	Score    int `json:"score"`
//	GameID   int `json:"game_id"`
//	PlayerID int `json:"player_id"`
//}

//func generatePlayers(numPlayers int) []Player {
//	players := make([]Player, numPlayers)
//	for i := 0; i < numPlayers; i++ {
//		players[i] = Player{i, gofakeit.Name()}
//	}
//	return players
//}

//func generateScore(numPlayers int) Score {
//	score := rand.Intn(20000)
//	gameId := rand.Intn(100)
//	playerId := rand.Intn(numPlayers - 1)
//	return Score{score, gameId, playerId}
//}

func push(writer *kafka.Writer, parent context.Context, key, value []byte, topic string) (err error) {
	message := kafka.Message{Topic: topic, Key: key, Value: value}
	return writer.WriteMessages(parent, message)
}

//func getScores(numPlayers int) <-chan Score {
//	s := make(chan Score)
//
//	go func() {
//		for {
//			s <- generateScore(numPlayers)
//			time.Sleep(50 * time.Millisecond)
//		}
//		close(s)
//	}()
//
//	return s
//}

//func getTimeEncoded() []byte {
//	return []byte(time.Now().UTC().Format("2006-01-02 15:04:05"))
//}

func main() {
//	numPlayers := 50
	numGames := len(snesGames)

// 	players := generatePlayers(numPlayers)

	w := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Async:        true,
		RequiredAcks: kafka.RequireNone,
	}
	defer w.Close()

// 	for i := 0; i < numPlayers; i++ {
// 		playerEncoded, err := json.Marshal(players[i])
// 		if err != nil {
// 			log.Println("failed to marshal player: ", err)
// 		}
// 		err = push(w, context.Background(), getTimeEncoded(), playerEncoded, "players")
// 		if err != nil {
// 			log.Println("failed to write message: ", err)
// 		}
// 	}

	for i := 0; i < numGames - 1; i++ {
        fmt.Println(" snesGames[",i,"] ==> ", snesGames[i])
		gameEncoded, err := json.Marshal(snesGames[i])
		if err != nil {
			log.Println("failed to marshal game: ", err)
		}
		err = push(w, context.Background(), []byte(strconv.Itoa(snesGames[i].Id)), gameEncoded, "snesgames")
		if err != nil {
			log.Println("failed to write message: ", err)
		}
	}

// 	scores := getScores(numPlayers)
// 
// 	for score := range scores {
// 		scoreEncoded, err := json.Marshal(score)
// 		if err != nil {
// 			log.Println("failed to marshal score: ", err)
// 		}
// 		err = push(w, context.Background(), getTimeEncoded(), scoreEncoded, "scores")
// 		if err != nil {
// 			log.Println("failed to write message: ", err)
// 		}
// 	}

}
