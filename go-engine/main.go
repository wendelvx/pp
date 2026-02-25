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
	Status          string         `json:"status"` // WAITING, BATTLE, GAMEOVER, VICTORY
	Boss            BossState      `json:"boss"`
	TeamHP          int            `json:"team_hp"`
	MaxTeamHP       int            `json:"max_team_hp"`
	Multiplier      float64        `json:"multiplier"`
	ActiveIncident  string         `json:"active_incident"`
	IncidentClicks  int            `json:"incident_clicks"`
	ClassCounts     map[string]int `json:"class_counts"` // Para validar início
	LastAttackAt    time.Time      `json:"-"`
	mu              sync.Mutex
}

var state = GameState{
	Status:      "WAITING",
	Boss:        BossState{HP: 1000000, MaxHP: 1000000},
	TeamHP:      1000,
	MaxTeamHP:   1000,
	Multiplier:  1.0,
	ClassCounts: make(map[string]int),
}

// --- MÉTRICAS ---
var (
	bossHPMetric = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dungeon_boss_hp"})
	teamHPMetric = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dungeon_team_hp"})
)

func main() {
	prometheus.MustRegister(bossHPMetric, teamHPMetric)
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})

	// 1. Endpoint de Métricas
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// 2. LOOP DE ATAQUE DO BOSS (Dano Automático)
	go func() {
		for {
			time.Sleep(3 * time.Second)
			state.mu.Lock()
			if state.Status == "BATTLE" {
				dano := 20
				// Se estiver em Legacy Code, o dano dobra
				if state.ActiveIncident == "legacy_code_spill" {
					dano = 50
				}
				state.TeamHP -= dano
				if state.TeamHP <= 0 {
					state.Status = "GAMEOVER"
				}
				teamHPMetric.Set(float64(state.TeamHP))
				broadcastState(rdb)
			}
			state.mu.Unlock()
		}
	}()

	// 3. GERADOR DE CAOS (Incidentes Randômicos Equilibrados)
	go func() {
		incidents := []string{"error_500", "code_review", "phishing", "database_lock", "legacy_code_spill"}
		for {
			time.Sleep(time.Duration(rand.Intn(20)+20) * time.Second)

			state.mu.Lock()
			if state.Status == "BATTLE" && state.ActiveIncident == "" {
				// Sorteio garantindo que não repita tanto o mesmo
				idx := rand.Intn(len(incidents))
				state.ActiveIncident = incidents[idx]
				state.IncidentClicks = 0
				log.Printf("🔥 EVENTO: %s\n", state.ActiveIncident)
				broadcastState(rdb)
			}
			state.mu.Unlock()
		}
	}()

	// 4. PROCESSAMENTO DE MENSAGENS
	pubsub := rdb.Subscribe(ctx, "player_attacks")
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			continue
		}

		go func(payload string) {
			var data map[string]interface{}
			json.Unmarshal([]byte(payload), &data)
			
			pClass := data["class"].(string)
			isJoin := data["type"] == "join" // Precisamos que o Node envie o type

			state.mu.Lock()
			defer state.mu.Unlock()

			// Lógica de Registro/Início
			if isJoin {
				state.ClassCounts[pClass]++
				checkStartGame()
			}

			if state.Status == "BATTLE" {
				if state.ActiveIncident != "" {
					processIncident(pClass)
				} else {
					processNormalAttack(pClass)
				}
			}

			broadcastState(rdb)
		}(msg.Payload)
	}
}

func checkStartGame() {
	if state.Status != "WAITING" {
		return
	}
	// Verifica se todas as 5 classes têm ao menos 1 player
	required := []string{"Front-end", "Back-end", "DevOps", "QA", "Security"}
	ready := 0
	for _, c := range required {
		if state.ClassCounts[c] > 0 {
			ready++
		}
	}
	if ready == 5 {
		state.Status = "BATTLE"
		log.Println("🚀 TODAS AS CLASSES PRESENTES. INICIANDO BATALHA!")
	}
}

func processNormalAttack(pClass string) {
	switch pClass {
	case "Front-end":
		dano := 300.0
		if rand.Float64() < 0.25 { dano *= 5 } // Crítico
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
			if state.IncidentClicks >= 20 { state.ActiveIncident = "" }
		} else {
			state.Boss.HP += 1000 // Penalidade por spam
		}
		state.LastAttackAt = time.Now()
	case "phishing":
		if pClass == "Security" {
			state.IncidentClicks++
			if state.IncidentClicks >= 20 { state.ActiveIncident = "" }
		}
	case "database_lock":
		// RF NOVO: Apenas Security destrava. Ninguém mais dá dano enquanto travado.
		if pClass == "Security" {
			state.IncidentClicks++
			if state.IncidentClicks >= 40 { state.ActiveIncident = "" }
		}
	case "legacy_code_spill":
		// RF NOVO: Back-end "refatora" para diminuir clicks necessários. 
		// QA precisa curar o dano extra que o boss está dando no loop automático.
		if pClass == "Back-end" {
			state.IncidentClicks++
			if state.IncidentClicks >= 50 { state.ActiveIncident = "" }
		}
	}
}

func broadcastState(rdb *redis.Client) {
	payload, _ := json.Marshal(state)
	rdb.Publish(ctx, "boss_updates", payload)
}