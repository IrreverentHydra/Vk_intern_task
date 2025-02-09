FROM node:16-alpine

WORKDIR /app

COPY package*.json ./

RUN if ! command -v npm > /dev/null; then \
      echo "npm not found"; exit 1; \
    fi && \
    npm install --production

COPY . .

EXPOSE 3000
CMD ["npm", "start"]
