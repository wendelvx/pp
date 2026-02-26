const { pub } = require('../config/redis');
const { attackCounter } = require('../config/metrics');

module.exports = (io, socket) => {
    // 1. Join
    socket.on('join_game', (data) => {
        socket.data.playerClass = data.class;
        socket.data.nickname = data.nickname;

        pub.publish('player_attacks', JSON.stringify({
            type: 'join',
            class: data.class,
            nickname: data.nickname,
            timestamp: Date.now()
        }));
    });

    // 2. Attack
    socket.on('attack', (data) => {
        const playerClass = socket.data.playerClass;
        if (!playerClass) return;

        attackCounter.inc({ class: playerClass });

        pub.publish('player_attacks', JSON.stringify({
            type: 'attack',
            class: playerClass,
            isFake: data?.isFake || false,
            timestamp: Date.now()
        }));
    });

    // 3. Admin Command
    socket.on('admin_start_game', () => {
        pub.publish('player_attacks', JSON.stringify({
            type: 'start_command',
            timestamp: Date.now()
        }));
    });
};