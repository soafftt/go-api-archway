# 로컬 인프라 실행 (.doker)

이 프로젝트는 로컬 실행 시 **PostgreSQL**과 **Valkey**가 필요합니다.

## 사전 준비

`docker compose`가 설치되어 있어야 합니다.

```bash
docker compose version
```

## 인프라 포트/컨테이너

| Service | Container | Host Port | Container Port |
|---|---|---:|---:|
| PostgreSQL | `archway-postgres` | `5431` | `5432` |
| Valkey master | `archway-valkey-master` | `6379` | `6379` |
| Valkey replica | `archway-valkey-replica` | `6380` | `6379` |

## 실행

```bash
cd .doker
docker compose up -d
```

## 중지

```bash
cd .doker
docker compose down
```

## console backend 연결 예시

```bash
POSTGRES_CONNECTION_STRING=postgres://postgres:postgres@127.0.0.1:5431/postgres
VALKEY_URL=redis://127.0.0.1:6379
```
