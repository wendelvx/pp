const { pub } = require('../config/redis');
const { attackCounter } = require('../config/metrics');

module.exports = (io, socket) => {
    // 1. Join (Entrada no Jogo)
    socket.on('join_game', (data) => {
        // Guardamos no estado da conexão do socket
        socket.data.playerClass = data.class;
        socket.data.nickname = data.nickname;

        pub.publish('player_attacks', JSON.stringify({
            type: 'join',
            class: data.class,
            nickname: data.nickname,
            timestamp: Date.now()
        }));
    });

    // 2. Attack (O clique do aluno)
    socket.on('attack', (data) => {
        const { playerClass, nickname } = socket.data;
        
        // Segurança: Se o aluno não "entrou" no jogo, o ataque é ignorado
        if (!playerClass || !nickname) return;

        // Incrementa métrica do Prometheus (Node)
        attackCounter.inc({ class: playerClass });

        // Envia para o Go Engine com o Nickname para o Ranking
        pub.publish('player_attacks', JSON.stringify({
            type: 'attack',
            class: playerClass,
            nickname: nickname, // CRUCIAL para a V2.0
            isFake: data?.isFake || false,
            timestamp: Date.now()
        }));
    });

    // 3. Admin Command (Início/Reset da Partida)
    socket.on('admin_start_game', (data) => {
        // Agora podemos escolher qual professor enfrentar via boss_id
        pub.publish('player_attacks', JSON.stringify({
            type: 'start_command',
            boss_id: data?.boss_id || 'tcc_titan', // 'arquiteto', 'pixel_perfect', etc.
            timestamp: Date.now()
        }));
    });
};