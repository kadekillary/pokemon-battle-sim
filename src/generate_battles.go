package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"

	"github.com/segmentio/kafka-go"
)

func checkErr(e error) {
	if e != nil {
		log.Println("error: ", e)
	}
}

type BattleDetails struct {
	Level int `json:"pokemon_level"`
	Id    int `json:"battle_id"`
}

var pokemonLen = len(pokemon)

func push(writer *kafka.Writer, parent context.Context, key, value []byte, topic string) (err error) {
	message := kafka.Message{Topic: topic, Key: key, Value: value}
	return writer.WriteMessages(parent, message)
}

func main() {
	w := &kafka.Writer{
		Addr:  kafka.TCP("localhost:9092"),
		Async: true,
	}
	defer w.Close()

	for i := 0; ; i++ {
		battleId := i
		attacker := []byte(pokemon[rand.Intn(pokemonLen-1)])
		attackerDetails, _ := json.Marshal(BattleDetails{rand.Intn(95) + 5, battleId})
        defender := []byte(pokemon[rand.Intn(pokemonLen-1)])
        defenderDetails, _ := json.Marshal(BattleDetails{rand.Intn(95) + 5, battleId})

        err := push(w, context.Background(), attacker, attackerDetails, "battle_attacker")
        checkErr(err)
        err = push(w, context.Background(), defender, defenderDetails, "battle_defender")
        checkErr(err)
	}
}
