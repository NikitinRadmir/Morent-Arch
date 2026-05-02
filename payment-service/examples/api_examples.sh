#!/usr/bin/env sh
set -e
BASE=${BASE:-http://localhost:8081/api/v1}

alice=$(curl -s -X POST "$BASE/accounts" -H 'Content-Type: application/json' -d '{"owner":"Alice","currency":"RUB"}' | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
bob=$(curl -s -X POST "$BASE/accounts" -H 'Content-Type: application/json' -d '{"owner":"Bob","currency":"RUB"}' | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

curl -s -X POST "$BASE/accounts/$alice/deposit" -H 'Content-Type: application/json' -d '{"amount_minor":10000,"currency":"RUB"}'
curl -s -X POST "$BASE/transfers" -H 'Content-Type: application/json' -H 'Idempotency-Key: demo-1' -d "{\"from_account_id\":\"$alice\",\"to_account_id\":\"$bob\",\"amount_minor\":2500,\"fee_minor\":100,\"currency\":\"RUB\"}"
curl -s "$BASE/accounts/$alice/summary"
