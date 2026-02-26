package game

import (
	"math/rand"
	"time"
	"dungeon-engine/internal/metrics"
	"github.com/go-redis/redis/v8"
)

func StartBossLoop(rdb *redis.Client) {
	for {
		time.Sleep(3 * time.Second)
		State.Mu.Lock()
		if State.Status == "BATTLE" {
			dano := 20
			if State.ActiveIncident == "legacy_code_spill" { dano = 50 }
			State.TeamHP -= dano
			if State.TeamHP <= 0 {
				State.TeamHP = 0
				State.Status = "GAMEOVER"
			}
			metrics.TeamHPMetric.Set(float64(State.TeamHP))
			BroadcastState(rdb)
		}
		State.Mu.Unlock()
	}
}

func StartChaosGenerator(rdb *redis.Client) {
	incidents := []string{"error_500", "code_review", "phishing", "database_lock", "legacy_code_spill"}
	for {
		time.Sleep(time.Duration(rand.Intn(15)+15) * time.Second)
		State.Mu.Lock()
		if State.Status == "BATTLE" && State.ActiveIncident == "" {
			State.ActiveIncident = incidents[rand.Intn(len(incidents))]
			State.IncidentClicks = 0
			BroadcastState(rdb)
		}
		State.Mu.Unlock()
	}
}