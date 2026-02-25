# Usamos uma versão leve do Node (Alpine) para performance
FROM node:18-alpine

# Define a pasta de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de dependência primeiro (otimiza o cache do Docker)
COPY package*.json ./

# Instala as dependências
RUN npm install --production

# Copia o resto do código (server.js)
COPY . .

# Expõe a porta que o Express está ouvindo
EXPOSE 3001

# Comando para iniciar o servidor
CMD ["node", "server.js"]