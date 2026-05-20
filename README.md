# 2030 KOREA StablePay Network

A stablecoin payment network portfolio project focused on low-fee merchant payments, reliable backend ledgers, on-chain settlement, and blockchain network infrastructure.

## Current Scope

Week 1 Day 1 starts with the public project foundation:

- Go HTTP API server
- `GET /health` endpoint
- PostgreSQL local development environment
- Repository structure for future merchant, invoice, and payment domains

## Project Structure

```text
cmd/api/                 API server entrypoint
internal/httpapi/         HTTP handlers and route registration
docker-compose.yml        Local PostgreSQL environment
```

## Run API

```bash
go run ./cmd/api
```

Health check:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok","service":"stablepay-api"}
```

## Run PostgreSQL

```bash
docker compose up -d
```

## Roadmap

The private study plan repository tracks detailed daily learning plans and retrospectives. This public repository contains only implementation artifacts intended for portfolio review.

## Project Structure

```text
cmd/api/                  API server entrypoint
internal/httpapi/          HTTP handlers and route registration
internal/platform/database PostgreSQL connection helper
migrations/                SQL migration files
docker-compose.yml         Local PostgreSQL environment
```

## Run PostgreSQL

```bash
docker compose up -d
docker compose ps
```

## Database Schema

The first schema contains three core tables:

- `merchants`: merchant profile
- `invoices`: payment request created by a merchant
- `payments`: payment attempt and on-chain transaction information


## Merchant API

Create merchant:

```bash
curl -i -X POST http://localhost:8080/merchants \
  -H "Content-Type: application/json" \
  -d '{"name":"Cafe Korea","email":"owner@cafe.example"}'
```

Expected status:

```text
201 Created
```