# syntax=docker/dockerfile:1
FROM index.docker.io/library/node@sha256:c48cf8c493930d6b5fbada793144b177113fefeda5397e99173938c59933285d
RUN apk add --no-cache python2 g++ make
WORKDIR /app
COPY . .
RUN yarn install --production
CMD ["node", "src/index.js"]
EXPOSE 3000
