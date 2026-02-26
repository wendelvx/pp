package game

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

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
	Mu             sync.Mutex     `json:"-"`
}

var State = GameState{
	Status:      "WAITING",
	Boss:        BossState{HP: 1000000, MaxHP: 1000000},
	TeamHP:      1000,
	MaxTeamHP:   1000,
	Multiplier:  1.0,
	ClassCounts: make(map[string]int),
}

func BroadcastState(rdb *redis.Client) {
	payload, _ := json.Marshal(State)
	rdb.Publish(context.Background(), "boss_updates", payload)
}