#!/bin/bash

# API Examples for Generator Service

BASE_URL="http://localhost:8080/api/v1"

echo "=== Generator Service API Examples ==="
echo ""

# Health check
echo "1. Health Check:"
curl -s "${BASE_URL%/api/v1}/health" | jq .
echo ""

# Generate password automatically
echo "2. Generate Password (Auto):"
curl -s "$BASE_URL/password" | jq .
echo ""

# Generate password by mask
echo "3. Generate Password by Mask (LLLdddss):"
curl -s "$BASE_URL/password/mask?mask=LLLdddss" | jq .
echo ""

# Generate QR code
echo "4. Generate QR Code:"
curl -s "$BASE_URL/qrcode?data=HelloWorld&size=256" --output qr_example.png
echo "QR code saved to qr_example.png"
echo ""

echo "=== Examples Complete ==="
