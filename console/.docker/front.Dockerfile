# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS deps
WORKDIR /workspace

COPY package.json package-lock.json ./
COPY front/package.json front/package.json
COPY backend/package.json backend/package.json

RUN npm ci --workspace @console/front --include-workspace-root=false

FROM deps AS build
WORKDIR /workspace

COPY front ./front
RUN npm run build --workspace @console/front

FROM scratch AS artifacts
COPY --from=build /workspace/front/dist/ /front/

FROM nginx:1.27-alpine AS runtime
COPY .docker/front.nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /workspace/front/dist /usr/share/nginx/html

EXPOSE 8081

CMD ["nginx", "-g", "daemon off;"]
