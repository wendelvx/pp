const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const Redis = require('ioredis');
const client = require('prom-client');

const app = express();
const server = http.createServer(app);
const io = new Server(server, { cors: { origin: "*" } });

// --- CONFIGURAÇÃO DE MÉTRICAS (PROMETHEUS) ---
const register = new client.Registry();
client.collectDefaultMetrics({ register });

const attackCounter = new client.Counter({
  name: 'dungeon_player_attacks_total',
  help: 'Total de ataques desferidos pelos alunos',
  labelNames: ['class']
});
register.registerMetric(attackCounter);

app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType);
  res.end(await register.metrics());
});

// --- CONFIGURAÇÃO DO REDIS ---
const redisPub = new Redis({ host: 'redis', port: 6379 });
const redisSub = new Redis({ host: 'redis', port: 6379 });

// Ouvir respostas da Engine Go e repassar para todos os alunos
redisSub.subscribe('boss_updates');
redisSub.on('message', (channel, message) => {
  // Broadcast do estado global do jogo (HP, Incidentes, Multiplicador, Status)
  io.emit('boss_update', JSON.parse(message));
});

// --- LÓGICA DE CONEXÃO DOS ALUNOS ---
io.on('connection', (socket) => {
  console.log(`Conexão estabelecida: ${socket.id}`);

  // Evento de Entrada (Join)
  socket.on('join_game', (data) => {
    // data = { nickname: 'Panda', class: 'Front-end' }
    socket.data.playerClass = data.class;
    socket.data.nickname = data.nickname;

    console.log(`[JOIN] ${data.nickname} escolheu ${data.class}`);

    // Notifica o Go para ele incrementar o contador de classes e checar se o jogo pode começar
    redisPub.publish('player_attacks', JSON.stringify({
      type: 'join', // O Go usará esse campo para diferenciar de um ataque
      class: data.class,
      nickname: data.nickname,
      timestamp: Date.now()
    }));
  });

  // Evento de Ação (Ataque ou Interação com Incidente)
  socket.on('attack', (data) => {
    const playerClass = socket.data.playerClass;
    
    // Se o player ainda não escolheu classe, ignora
    if (!playerClass) return;

    // 1. Incrementa métrica para o Grafana (RNF03)
    attackCounter.inc({ class: playerClass });

    // 2. Envia para o Redis
    redisPub.publish('player_attacks', JSON.stringify({
      type: 'attack',
      class: playerClass,
      isFake: data?.isFake || false, // Importante para o Phishing do Security
      timestamp: Date.now()
    }));
  });

  socket.on('disconnect', () => {
    console.log(`Usuário desconectado: ${socket.id}`);
  });
});

server.listen(3001, () => {
  console.log('🚀 Gateway Node.js (Porta 3001) aguardando conexões...');
});