package models

import "time"

// BossProfile define os atributos variáveis de cada professor
type BossProfile struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	BaseHP       int           `json:"base_hp"`
	BaseDamage   int           `json:"base_damage"`
	AttackSpeed  time.Duration `json:"attack_speed"`
	IncidentBias string        `json:"incident_bias"`
}

// PlayerStats armazena a performance individual para o ranking
type PlayerStats struct {
	Nickname        string `json:"nickname"`
	Class           string `json:"class"`
	TotalDamage     int    `json:"total_damage"`
	IncidentsSolved int    `json:"incidents_solved"`
}

// IncidentMeta define o estado do incidente atual
type IncidentMeta struct {
	ID            string `json:"id"`
	TargetQuota   int    `json:"target_quota"`
	CurrentClicks int    `json:"current_clicks"`
	RequiredClass string `json:"required_class"`
}