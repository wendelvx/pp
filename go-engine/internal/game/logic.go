package game

import (
	"log"
	"math/rand"
	"time"
	"dungeon-engine/internal/metrics"
)

func CheckStartGame() {
	if State.Status != "WAITING" { return }
	required := []string{"Front-end", "Back-end", "DevOps", "QA", "Security"}
	ready := 0
	for _, c := range required {
		if State.ClassCounts[c] > 0 { ready++ }
	}
	if ready == 5 { State.Status = "BATTLE" }
}

func ResetGame() {
	State.Status = "WAITING"
	State.ClassCounts = make(map[string]int)
	State.ResetTrigger++
	State.Boss.HP = State.Boss.MaxHP
	State.TeamHP = State.MaxTeamHP
	State.Multiplier = 1.0
	State.ActiveIncident = ""
	State.IncidentClicks = 0
	
	metrics.BossHPMetric.Set(float64(State.Boss.HP))
	metrics.TeamHPMetric.Set(float64(State.TeamHP))
	log.Println("♻️ SISTEMA RESETADO: Voltando para seleção...")
}

func ProcessNormalAttack(pClass string) {
	switch pClass {
	case "Front-end":
		dano := 300.0
		if rand.Float64() < 0.25 { dano *= 5 }
		State.Boss.HP -= int(dano * State.Multiplier)
	case "Back-end":
		State.Boss.HP -= int(400 * State.Multiplier)
	case "DevOps":
		State.Multiplier += 0.01
	case "QA":
		if State.TeamHP < State.MaxTeamHP { State.TeamHP += 20 }
	case "Security":
		State.Boss.HP -= int(250 * State.Multiplier)
	}
	metrics.BossHPMetric.Set(float64(State.Boss.HP))
}

func ProcessIncident(pClass string) {
	switch State.ActiveIncident {
	case "error_500":
		if pClass == "DevOps" {
			State.IncidentClicks++
			if State.IncidentClicks >= 30 { State.ActiveIncident = "" }
		}
	case "code_review":
		if time.Since(State.LastAttackAt) > 800*time.Millisecond {
			State.Boss.HP -= 2000
			State.IncidentClicks++
		} else {
			State.Boss.HP += 5000 
		}
		if State.IncidentClicks >= 20 { State.ActiveIncident = "" }
		State.LastAttackAt = time.Now()
	case "phishing":
		if pClass == "Security" {
			State.IncidentClicks++
			if State.IncidentClicks >= 20 { State.ActiveIncident = "" }
		}
	case "database_lock":
		if pClass == "Security" {
			State.IncidentClicks++
			if State.IncidentClicks >= 40 { State.ActiveIncident = "" }
		}
	case "legacy_code_spill":
		if pClass == "Back-end" {
			State.IncidentClicks++
			if State.IncidentClicks >= 50 { State.ActiveIncident = "" }
		}
	}
	metrics.BossHPMetric.Set(float64(State.Boss.HP))
}