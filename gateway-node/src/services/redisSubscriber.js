const { sub } = require('../config/redis');

module.exports = (io) => {
    sub.subscribe('boss_updates');
    
    sub.on('message', (channel, message) => {
        if (channel === 'boss_updates') {
            try {
                const gameState = JSON.parse(message);
                
                // Emite para todos os celulares conectados
                // O payload agora contém: boss, status, team_hp, active_incident (com quotas), etc.
                io.emit('boss_update', gameState);
            } catch (err) {
                console.error("❌ Erro ao processar boss_update do Go:", err);
            }
        }
    });
};