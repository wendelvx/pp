package game

import (
	"log"
	"math/rand"
	"time"

	"dungeon-engine/internal/db"
	"dungeon-engine/internal/metrics"
	"dungeon-engine/internal/models" // Importação vital para as structs
	"github.com/go-redis/redis/v8"
)

func StartBossLoop(rdb *redis.Client) {
	for {
		// O intervalo de ataque vem do Perfil do Boss definido no models.BossProfile
		time.Sleep(State.Boss.AttackSpeed)

		State.Mu.Lock()
		if State.Status == "BATTLE" {
			dano := State.Boss.BaseDamage

			if State.ActiveIncident.ID == "legacy_code_spill" {
				dano *= 2
			}

			State.TeamHP -= dano

			if State.TeamHP <= 0 {
				State.TeamHP = 0
				State.Status = "GAMEOVER"

				// Persistência assíncrona usando o pacote db e models
				db.SaveBattleResult(
					State.Status,
					State.Boss.ID,
					State.StartTime,
					State.Players,
				)

				log.Printf("💀 DERROTA: O Boss %s venceu a turma.", State.Boss.Name)
			}

			metrics.TeamHPMetric.Set(float64(State.TeamHP))
			BroadcastState(rdb)
		}
		State.Mu.Unlock()
	}
}

func StartChaosGenerator(rdb *redis.Client) {
	// Mapa atualizado para usar a struct centralizada em models
	incidentTypes := map[string]models.IncidentMeta{
		"error_500":         {ID: "error_500", TargetQuota: 15, RequiredClass: "DevOps"},
		"code_review":       {ID: "code_review", TargetQuota: 20, RequiredClass: "All"},
		"phishing":          {ID: "phishing", TargetQuota: 20, RequiredClass: "Security"},
		"database_lock":     {ID: "database_lock", TargetQuota: 40, RequiredClass: "Security"},
		"legacy_code_spill": {ID: "legacy_code_spill", TargetQuota: 50, RequiredClass: "Back-end"},
	}

	keys := make([]string, 0, len(incidentTypes))
	for k := range incidentTypes {
		keys = append(keys, k)
	}

	for {
		time.Sleep(time.Duration(rand.Intn(15)+15) * time.Second)

		State.Mu.Lock()
		if State.Status == "BATTLE" && State.ActiveIncident.ID == "" {
			randomKey := keys[rand.Intn(len(keys))]
			State.ActiveIncident = incidentTypes[randomKey]

			log.Printf("🔥 CAOS INSTALADO: %s iniciado!\n", State.ActiveIncident.ID)
			BroadcastState(rdb)
		}
		State.Mu.Unlock()
	}
}