#!/bin/bash
# Basic connectivity test
echo "=== Admin API Health ==="
curl -v http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@saas.com","password":"admin123"}' 2>&1

echo ""
echo "=== Tenant API Health ==="
curl -v http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"joao@minha-loja.com","password":"senha12345","current_tenant_code":"minha-loja"}' 2>&1
