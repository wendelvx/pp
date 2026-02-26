package game

import (
	"log"
	"math/rand"
	"time"
	"dungeon-engine/internal/metrics"
	"dungeon-engine/internal/models" // Importação vital
)

func CheckStartGame() {
	if State.Status != "WAITING" { return }
	required := []string{"Front-end", "Back-end", "DevOps", "QA", "Security"}
	ready := 0
	for _, c := range required {
		if State.ClassCounts[c] > 0 { ready++ }
	}
	if ready == 5 { 
		State.Status = "BATTLE" 
		State.StartTime = time.Now()
		State.LastAttackAt = time.Now()
	}
}

func ResetGame() {
	State.Status = "WAITING"
	State.ClassCounts = make(map[string]int)
	// Referência ao models.PlayerStats
	State.Players = make(map[string]*models.PlayerStats)
	State.ResetTrigger++
	
	State.BossHP = State.Boss.BaseHP
	State.TeamHP = State.MaxTeamHP
	State.Multiplier = 1.0
	// Referência ao models.IncidentMeta
	State.ActiveIncident = models.IncidentMeta{} 
	State.LastAttackAt = time.Now() 
	
	metrics.BossHPMetric.Set(float64(State.BossHP))
	metrics.TeamHPMetric.Set(float64(State.TeamHP))
	
	log.Printf("♻️ SISTEMA RESETADO: Perfil ativo: %s\n", State.Boss.Name)
}

func ProcessNormalAttack(pClass string, nickname string) {
	dano := 0.0
	switch pClass {
	case "Front-end":
		dano = 300.0
		if rand.Float64() < 0.25 { dano *= 5 } 
	case "Back-end":
		dano = 400.0
	case "DevOps":
		State.Multiplier += 0.01 
	case "QA":
		if State.TeamHP < State.MaxTeamHP { State.TeamHP += 20 }
	case "Security":
		dano = 250.0
	}

	totalDano := int(dano * State.Multiplier)
	State.BossHP -= totalDano

	if p, ok := State.Players[nickname]; ok {
		p.TotalDamage += totalDano
	} else {
		// Instanciando models.PlayerStats
		State.Players[nickname] = &models.PlayerStats{
			Nickname: nickname,
			Class: pClass,
			TotalDamage: totalDano,
		}
	}
	metrics.BossHPMetric.Set(float64(State.BossHP))
}

func ProcessIncident(pClass string, nickname string) {
	if State.ActiveIncident.RequiredClass != "All" && State.ActiveIncident.RequiredClass != pClass {
		return 
	}

	if State.ActiveIncident.ID == "code_review" {
		if time.Since(State.LastAttackAt) < 800*time.Millisecond {
			State.BossHP += 5000 
			State.LastAttackAt = time.Now()
			return
		}
	}

	State.ActiveIncident.CurrentClicks++
	State.LastAttackAt = time.Now()

	if p, ok := State.Players[nickname]; ok {
		p.IncidentsSolved++
	}

	if State.ActiveIncident.CurrentClicks >= State.ActiveIncident.TargetQuota {
		// Resetando para uma struct models.IncidentMeta vazia
		State.ActiveIncident = models.IncidentMeta{} 
		log.Println("✅ Incidente resolvido!")
	}
	metrics.BossHPMetric.Set(float64(State.BossHP))
}