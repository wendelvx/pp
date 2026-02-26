package game

// Removido o import de models para evitar o erro "imported and not used"
// O Go não perdoa importação sobrando!

type Incident interface {
	GetName() string
	GetTargetQuota() int
	GetRequiredClass() string
	OnAttack(pClass string, state *GameState) bool 
}

type EmergencyDeploy struct{}

func (e *EmergencyDeploy) GetName() string          { return "error_500" }
func (e *EmergencyDeploy) GetTargetQuota() int      { return 15 }
func (e *EmergencyDeploy) GetRequiredClass() string { return "DevOps" }

func (e *EmergencyDeploy) OnAttack(pClass string, s *GameState) bool {
	// Retorna true se a classe do jogador tiver permissão para resolver este incidente
	return pClass == e.GetRequiredClass()
}