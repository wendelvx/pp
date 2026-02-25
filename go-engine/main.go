package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ctx = context.Background()

// --- ESTRUTURAS DE ESTADO ---

type BossState struct {
	HP    int `json:"hp"`
	MaxHP int `json:"max_hp"`
}

type GameState struct {
	Status         string         `json:"status"`
	Boss           BossState      `json:"boss"`
	TeamHP         int            `json:"team_hp"`
	MaxTeamHP      int            `json:"max_team_hp"`
	Multiplier     float64        `json:"multiplier"`
	ActiveIncident string         `json:"active_incident"`
	IncidentClicks int            `json:"incident_clicks"`
	ClassCounts    map[string]int `json:"class_counts"`
	ResetTrigger   int            `json:"reset_trigger"`
	LastAttackAt   time.Time      `json:"-"`
	mu             sync.Mutex
}

var state = GameState{
	Status:      "WAITING",
	Boss:        BossState{HP: 1000000, MaxHP: 1000000},
	TeamHP:      1000,
	MaxTeamHP:   1000,
	Multiplier:  1.0,
	ClassCounts: make(map[string]int),
	ResetTrigger: 0,
}

var (
	bossHPMetric = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dungeon_boss_hp"})
	teamHPMetric = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dungeon_team_hp"})
)

func main() {
	prometheus.MustRegister(bossHPMetric, teamHPMetric)
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// 2. LOOP DE ATAQUE DO BOSS
	go func() {
		for {
			time.Sleep(3 * time.Second)
			state.mu.Lock()
			if state.Status == "BATTLE" {
				dano := 20
				if state.ActiveIncident == "legacy_code_spill" { dano = 50 }
				state.TeamHP -= dano
				if state.TeamHP <= 0 {
					state.TeamHP = 0
					state.Status = "GAMEOVER"
				}
				teamHPMetric.Set(float64(state.TeamHP))
				broadcastState(rdb)
			}
			state.mu.Unlock()
		}
	}()

	// 3. GERADOR DE CAOS
	go func() {
		incidents := []string{"error_500", "code_review", "phishing", "database_lock", "legacy_code_spill"}
		for {
			time.Sleep(time.Duration(rand.Intn(15)+15) * time.Second)
			state.mu.Lock()
			if state.Status == "BATTLE" && state.ActiveIncident == "" {
				state.ActiveIncident = incidents[rand.Intn(len(incidents))]
				state.IncidentClicks = 0
				broadcastState(rdb)
			}
			state.mu.Unlock()
		}
	}()

	// 4. PROCESSAMENTO DE MENSAGENS
	pubsub := rdb.Subscribe(ctx, "player_attacks")
	for {
		msg, _ := pubsub.ReceiveMessage(ctx)
		go func(payload string) {
			var data map[string]interface{}
			json.Unmarshal([]byte(payload), &data)

			state.mu.Lock()
			defer state.mu.Unlock()

			msgType, _ := data["type"].(string)

			// --- LÓGICA INTELIGENTE DE START / RESET ---
			if msgType == "start_command" {
				if state.Status == "WAITING" {
					// SE ESTÁ NO LOBBY -> INICIA A BATALHA
					state.Status = "BATTLE"
					log.Println("⚔️ BATALHA INICIADA!")
				} else {
					// SE JÁ ACABOU (VICTORY/GAMEOVER) -> RESET TOTAL PARA O LOBBY
					state.Status = "WAITING"
					state.ClassCounts = make(map[string]int) 
					state.ResetTrigger++ // Notifica o mobile para limpar a classe
					state.Boss.HP = state.Boss.MaxHP
					state.TeamHP = state.MaxTeamHP
					state.Multiplier = 1.0
					state.ActiveIncident = ""
					state.IncidentClicks = 0
					log.Println("♻️ SISTEMA RESETADO: Voltando para seleção...")
				}
				
				bossHPMetric.Set(float64(state.Boss.HP))
				teamHPMetric.Set(float64(state.TeamHP))
				broadcastState(rdb)
				return
			}

			pClass, _ := data["class"].(string)
			isFake, _ := data["isFake"].(bool)

			if msgType == "join" {
				state.ClassCounts[pClass]++
				checkStartGame() 
			} else if state.Status == "BATTLE" && msgType == "attack" {
				if isFake {
					state.TeamHP -= 50 
				} else {
					if state.ActiveIncident != "" {
						processIncident(pClass)
					} else {
						processNormalAttack(pClass)
					}
				}
				if state.Boss.HP <= 0 {
					state.Boss.HP = 0
					state.Status = "VICTORY"
				}
			}
			broadcastState(rdb)
		}(msg.Payload)
	}
}

// ... (Mantenha as funções checkStartGame, processNormalAttack, processIncident e broadcastState iguais)

func checkStartGame() {
	if state.Status != "WAITING" { return }
	required := []string{"Front-end", "Back-end", "DevOps", "QA", "Security"}
	ready := 0
	for _, c := range required {
		if state.ClassCounts[c] > 0 { ready++ }
	}
	if ready == 5 { state.Status = "BATTLE" }
}

func processNormalAttack(pClass string) {
	switch pClass {
	case "Front-end":
		dano := 300.0
		if rand.Float64() < 0.25 { dano *= 5 }
		state.Boss.HP -= int(dano * state.Multiplier)
	case "Back-end":
		state.Boss.HP -= int(400 * state.Multiplier)
	case "DevOps":
		state.Multiplier += 0.01
	case "QA":
		if state.TeamHP < state.MaxTeamHP { state.TeamHP += 20 }
	case "Security":
		state.Boss.HP -= int(250 * state.Multiplier)
	}
	bossHPMetric.Set(float64(state.Boss.HP))
}

func processIncident(pClass string) {
	switch state.ActiveIncident {
	case "error_500":
		if pClass == "DevOps" {
			state.IncidentClicks++
			if state.IncidentClicks >= 30 { state.ActiveIncident = "" }
		}
	case "code_review":
		if time.Since(state.LastAttackAt) > 800*time.Millisecond {
			state.Boss.HP -= 2000
			state.IncidentClicks++
		} else {
			state.Boss.HP += 5000 
		}
		if state.IncidentClicks >= 20 { state.ActiveIncident = "" }
		state.LastAttackAt = time.Now()
	case "phishing":
		if pClass == "Security" {
			state.IncidentClicks++
			if state.IncidentClicks >= 20 { state.ActiveIncident = "" }
		}
	case "database_lock":
		if pClass == "Security" {
			state.IncidentClicks++
			if state.IncidentClicks >= 40 { state.ActiveIncident = "" }
		}
	case "legacy_code_spill":
		if pClass == "Back-end" {
			state.IncidentClicks++
			if state.IncidentClicks >= 50 { state.ActiveIncident = "" }
		}
	}
	bossHPMetric.Set(float64(state.Boss.HP))
}

func broadcastState(rdb *redis.Client) {
	payload, _ := json.Marshal(state)
	rdb.Publish(ctx, "boss_updates", payload)
}