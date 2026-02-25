const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const Redis = require('ioredis');
const client = require('prom-client');

const app = express();
const server = http.createServer(app);
const io = new Server(server, { 
    cors: { origin: "*" } 
});

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
// Conecta ao container 'redis' definido no docker-compose
const redisPub = new Redis({ host: 'redis', port: 6379 });
const redisSub = new Redis({ host: 'redis', port: 6379 });

// Ouvir respostas da Engine Go e repassar para todos os clientes (App e Central)
redisSub.subscribe('boss_updates');
redisSub.on('message', (channel, message) => {
  io.emit('boss_update', JSON.parse(message));
});

// --- LÓGICA DE CONEXÃO ---
io.on('connection', (socket) => {
  console.log(`Conexão estabelecida: ${socket.id}`);

  // 1. Evento de Entrada (Join do Aluno)
  socket.on('join_game', (data) => {
    socket.data.playerClass = data.class;
    socket.data.nickname = data.nickname;

    console.log(`[JOIN] ${data.nickname} escolheu ${data.class}`);

    // Publica no Redis para o Go registrar a entrada
    redisPub.publish('player_attacks', JSON.stringify({
      type: 'join',
      class: data.class,
      nickname: data.nickname,
      timestamp: Date.now()
    }));
  });

  // 2. Evento de Ação (Ataque do Aluno)
  socket.on('attack', (data) => {
    const playerClass = socket.data.playerClass;
    if (!playerClass) return;

    // Incrementa métrica para o Grafana
    attackCounter.inc({ class: playerClass });

    // Envia o ataque para processamento na Engine Go
    redisPub.publish('player_attacks', JSON.stringify({
      type: 'attack',
      class: playerClass,
      isFake: data?.isFake || false,
      timestamp: Date.now()
    }));
  });

  // 3. NOVO: Comando do Mestre (Botão Start no index.html)
  // Este evento é disparado pela Central de Comando para forçar o início
  socket.on('admin_start_game', () => {
    console.log("🕹️  Comando de Administrador: Iniciando Batalha!");
    
    redisPub.publish('player_attacks', JSON.stringify({
      type: 'start_command',
      timestamp: Date.now()
    }));
  });

  socket.on('disconnect', () => {
    console.log(`Usuário desconectado: ${socket.id}`);
  });
});

// Porta 3001 para bater com o docker-compose e o Cloudflare Tunnel
server.listen(3001, () => {
  console.log('🚀 Gateway Node.js Online (Porta 3001)');
});