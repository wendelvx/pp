package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"dungeon-engine/internal/game"
	"dungeon-engine/internal/metrics"
	"dungeon-engine/internal/redis" // Seu novo package

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()
	metrics.Init()
	rdb := redis.NewClient("redis:6379")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	go game.StartBossLoop(rdb)
	go game.StartChaosGenerator(rdb)

	pubsub := rdb.Subscribe(ctx, "player_attacks")
	log.Println("⚔️ Dungeon Engine Online!")

	for {
		msg, _ := pubsub.ReceiveMessage(ctx)
		go func(payload string) {
			var data map[string]interface{}
			json.Unmarshal([]byte(payload), &data)

			game.State.Mu.Lock()
			defer game.State.Mu.Unlock()

			msgType, _ := data["type"].(string)

			// Lógica de Start/Reset
			if msgType == "start_command" {
				if game.State.Status == "WAITING" {
					game.State.Status = "BATTLE"
					log.Println("⚔️ BATALHA INICIADA!")
				} else {
					game.ResetGame()
				}
				game.BroadcastState(rdb)
				return
			}

			// Lógica de Jogo
			pClass, _ := data["class"].(string)
			isFake, _ := data["isFake"].(bool)

			if msgType == "join" {
				game.State.ClassCounts[pClass]++
				game.CheckStartGame()
			} else if game.State.Status == "BATTLE" && msgType == "attack" {
				if isFake {
					game.State.TeamHP -= 50
					metrics.TeamHPMetric.Set(float64(game.State.TeamHP))
				} else {
					if game.State.ActiveIncident != "" {
						game.ProcessIncident(pClass)
					} else {
						game.ProcessNormalAttack(pClass)
					}
				}
				if game.State.Boss.HP <= 0 {
					game.State.Boss.HP = 0
					game.State.Status = "VICTORY"
				}
			}
			game.BroadcastState(rdb)
		}(msg.Payload)
	}
}