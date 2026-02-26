package db

import (
	"database/sql"
	"dungeon-engine/internal/models" // Importação corrigida
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(connStr string) {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Erro ao configurar driver Postgres: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Erro ao conectar ao Postgres: %v", err)
	}

	log.Println("✅ Conexão com Postgres estabelecida com sucesso!")
}

// Agora recebe o mapa vindo de models
func SaveBattleResult(status string, bossID string, startTime time.Time, players map[string]*models.PlayerStats) {
	// Snapshot usando models.PlayerStats
	playerSnapshot := make([]models.PlayerStats, 0, len(players))
	for _, p := range players {
		playerSnapshot = append(playerSnapshot, *p)
	}

	go func(stats []models.PlayerStats) {
		duration := time.Since(startTime).Seconds()
		
		var battleID int
		err := DB.QueryRow(
			"INSERT INTO battles (boss_id, result, duration) VALUES ($1, $2, $3) RETURNING id",
			bossID, status, int(duration),
		).Scan(&battleID)

		if err != nil {
			log.Printf("❌ Erro crítico ao salvar batalha: %v", err)
			return
		}

		tx, err := DB.Begin()
		if err != nil {
			log.Printf("❌ Erro ao iniciar transação: %v", err)
			return
		}

		for _, p := range stats {
			_, err := tx.Exec(
				`INSERT INTO rankings (nickname, class, total_damage, incidents_solved, battle_id) 
				 VALUES ($1, $2, $3, $4, $5)`,
				p.Nickname, p.Class, p.TotalDamage, p.IncidentsSolved, battleID,
			)
			if err != nil {
				log.Printf("⚠️ Erro ao inserir ranking de %s: %v", p.Nickname, err)
				tx.Rollback()
				return
			}
		}

		if err := tx.Commit(); err != nil {
			log.Printf("❌ Erro ao commitar transação: %v", err)
		} else {
			log.Printf("💾 Batalha #%d (%s) salva com %d jogadores!", battleID, status, len(stats))
		}
	}(playerSnapshot)
}