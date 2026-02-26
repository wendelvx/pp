package game

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	"dungeon-engine/internal/models" // Importação nova

	"github.com/go-redis/redis/v8"
)

type GameState struct {
	Status         string                          `json:"status"`
	Boss           models.BossProfile              `json:"boss"`
	BossHP         int                             `json:"boss_hp"`
	TeamHP         int                             `json:"team_hp"`
	MaxTeamHP      int                             `json:"max_team_hp"`
	Multiplier     float64                         `json:"multiplier"`
	ActiveIncident models.IncidentMeta             `json:"active_incident"`
	ClassCounts    map[string]int                  `json:"class_counts"`
	Players        map[string]*models.PlayerStats  `json:"players"`
	ResetTrigger   int                             `json:"reset_trigger"`
	StartTime      time.Time                       `json:"start_time"`
	LastAttackAt   time.Time                       `json:"-"`
	Mu             sync.Mutex                      `json:"-"`
}

var State = GameState{
	Status: "WAITING",
	Boss: models.BossProfile{
		ID:           "arquiteto",
		Name:         "O Arquiteto (Padrão)",
		BaseHP:       1000000,
		BaseDamage:   20,
		AttackSpeed:  3 * time.Second,
		IncidentBias: "database_lock",
	},
	TeamHP:      1000,
	MaxTeamHP:   1000,
	Multiplier:  1.0,
	ClassCounts: make(map[string]int),
	Players:     make(map[string]*models.PlayerStats),
}

func BroadcastState(rdb *redis.Client) {
	payload, _ := json.Marshal(State)
	rdb.Publish(context.Background(), "boss_updates", payload)
}