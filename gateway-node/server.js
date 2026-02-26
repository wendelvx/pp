const server = require('./src/app');
const PORT = process.env.PORT || 3001;

server.listen(PORT, () => {
    console.log(`
    =========================================
    🚀 GATEWAY DUNGEON MASTER ONLINE
    Porta: ${PORT}
    Status: Pronto para o Caos
    =========================================
    `);
});