const client = require('prom-client');

const register = new client.Registry();
client.collectDefaultMetrics({ register });

const attackCounter = new client.Counter({
    name: 'dungeon_player_attacks_total',
    help: 'Total de ataques desferidos pelos alunos',
    labelNames: ['class']
});

register.registerMetric(attackCounter);

module.exports = { register, attackCounter };