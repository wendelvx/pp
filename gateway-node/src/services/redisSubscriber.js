const { sub } = require('../config/redis');

module.exports = (io) => {
    sub.subscribe('boss_updates');
    
    sub.on('message', (channel, message) => {
        if (channel === 'boss_updates') {
            io.emit('boss_update', JSON.parse(message));
        }
    });
};