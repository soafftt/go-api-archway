# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS deps
WORKDIR /workspace

COPY package.json package-lock.json ./
COPY backend/package.json backend/package.json
COPY front/package.json front/package.json

RUN npm ci --workspace @console/backend --include-workspace-root=false

FROM deps AS build
WORKDIR /workspace

COPY backend ./backend
RUN npm run build --workspace @console/backend

FROM scratch AS artifacts
COPY --from=build /workspace/backend/dist/ /dist/
COPY --from=build /workspace/backend/sql/ /sql/
COPY --from=build /workspace/backend/package.json /package.json

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

COPY package.json package-lock.json ./
COPY backend/package.json backend/package.json
COPY front/package.json front/package.json
RUN npm ci --workspace @console/backend --include-workspace-root=false --omit=dev

COPY --from=build /workspace/backend/dist ./backend/dist
COPY --from=build /workspace/backend/sql ./backend/sql

WORKDIR /app/backend
EXPOSE 8080

CMD ["sh", "-c", "node dist/scripts/init-db.js && node dist/main.js"]
