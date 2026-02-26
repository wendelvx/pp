package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	BossHPMetric = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dungeon_boss_hp"})
	TeamHPMetric = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dungeon_team_hp"})
)

func Init() {
	prometheus.MustRegister(BossHPMetric, TeamHPMetric)
}