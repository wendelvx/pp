const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const { register } = require('./config/metrics');
const registerGameHandlers = require('./sockets/gameHandlers');
const initRedisSubscriber = require('./services/redisSubscriber');

const app = express();
const server = http.createServer(app);
const io = new Server(server, { cors: { origin: "*" } });

// Rota de Métricas
app.get('/metrics', async (req, res) => {
    res.set('Content-Type', register.contentType);
    res.end(await register.metrics());
});

// Inicializa o ouvinte do Redis (Go -> Node -> App)
initRedisSubscriber(io);

// Inicializa os eventos de Socket (App -> Node -> Go)
io.on('connection', (socket) => {
    console.log(`📡 Socket conectado: ${socket.id}`);
    registerGameHandlers(io, socket);

    socket.on('disconnect', () => console.log(`❌ Desconectado: ${socket.id}`));
});

module.exports = server;