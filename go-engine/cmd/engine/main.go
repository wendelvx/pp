package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"dungeon-engine/internal/boss"
	"dungeon-engine/internal/db"
	"dungeon-engine/internal/game"
	"dungeon-engine/internal/metrics"
	"dungeon-engine/internal/redis"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()

	// 1. Inicialização de Dependências
	metrics.Init()
	db.InitDB("postgres://admin:tcc_password@postgres:5432/dungeon_master?sslmode=disable")
	rdb := redis.NewClient("redis:6379")

	// 2. Servidor de Métricas (Prometheus)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("📊 Métricas disponíveis em :8080/metrics")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("❌ Erro no servidor de métricas:", err)
		}
	}()

	// 3. Loops de Processamento em Background
	go game.StartBossLoop(rdb)
	go game.StartChaosGenerator(rdb)

	// 4. Inscrição no Canal de Ataques do Redis
	pubsub := rdb.Subscribe(ctx, "player_attacks")
	log.Println("⚔️ Dungeon Engine V2.0 Online e Persistente!")

	// 5. Loop Principal de Eventos
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("⚠️ Erro ao receber mensagem do Redis: %v", err)
			continue
		}

		// Processamento assíncrono de cada mensagem para não travar o loop principal
		go func(payload string) {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(payload), &data); err != nil {
				log.Printf("❌ Erro ao decodificar JSON: %v", err)
				return
			}

			// --- SEÇÃO CRÍTICA: Trava o Mutex para garantir thread-safety ---
			game.State.Mu.Lock()
			
			msgType, _ := data["type"].(string)

			if msgType == "start_command" {
				// Lógica de Início/Reset da Partida
				if game.State.Status == "WAITING" {
					bossID, _ := data["boss_id"].(string)
					if bossID == "" { bossID = "tcc_titan" }
					
					game.State.Boss = boss.GetBossProfile(bossID)
					game.State.BossHP = game.State.Boss.BaseHP
					game.State.Status = "BATTLE"
					game.State.StartTime = time.Now()
					log.Printf("🚀 BATALHA INICIADA: %s", game.State.Boss.Name)
				} else {
					game.ResetGame()
				}
			} else {
				// Lógica de Interação dos Jogadores
				pClass, _ := data["class"].(string)
				nickname, _ := data["nickname"].(string)
				isFake, _ := data["isFake"].(bool)

				if msgType == "join" {
					game.State.ClassCounts[pClass]++
					game.CheckStartGame()
				} else if game.State.Status == "BATTLE" && msgType == "attack" {
					if isFake {
						// Punição para cliques em traps (Phishing)
						game.State.TeamHP -= 50
					} else {
						// Processa ataque normal ou resolução de incidente
						if game.State.ActiveIncident.ID != "" {
							game.ProcessIncident(pClass, nickname)
						} else {
							game.ProcessNormalAttack(pClass, nickname)
						}
					}

					// Verificação de Condição de Vitória
					if game.State.BossHP <= 0 {
						game.State.BossHP = 0
						game.State.Status = "VICTORY"
						
						// Salva o resultado final no Postgres de forma assíncrona
						db.SaveBattleResult(
							game.State.Status, 
							game.State.Boss.ID, 
							game.State.StartTime, 
							game.State.Players,
						)
						log.Println("🏆 VITÓRIA: O Boss foi derrotado!")
					}
				}
			}
			
			// Transmite o novo estado para o Node Gateway e LIBERA o Mutex
			game.BroadcastState(rdb)
			game.State.Mu.Unlock()
			// --- FIM DA SEÇÃO CRÍTICA ---
			
		}(msg.Payload)
	}
}