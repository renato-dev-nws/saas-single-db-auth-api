.PHONY: create-test-image test-product-image test-service-image \
        test-image-worker-check test-images-complete

# Creates a valid 10x10 PNG test file at test-image.jpg using Python
create-test-image:
	@python3 scripts/gen_test_image.py test-image.jpg

# Upload image to a product and verify worker processed all variants
test-product-image:
	@echo "========================================="
	@echo "Test: Product Image Upload + Worker"
	@echo "========================================="
	@if [ ! -f test-image.jpg ]; then $(MAKE) create-test-image; fi
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Logged in: URL_CODE=$$URL_CODE"; \
	PROD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Produto Imagem Teste","price":49.90,"stock":10,"is_active":true}'); \
	PRODUCT_ID=$$(echo "$$PROD" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "2. Product created: $$PRODUCT_ID"; \
	UPLOAD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID/image \
		-H "Authorization: Bearer $$TOKEN" \
		-F "image=@test-image.jpg;type=image/png"); \
	echo "3. Upload response: $$UPLOAD"; \
	echo "4. Waiting 4s for worker..."; \
	sleep 4; \
	COUNT=$$(docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -t -c \
		"SELECT COUNT(*) FROM images WHERE imageable_id='$$PRODUCT_ID' AND processing_status='completed';"); \
	echo "5. Completed variants: $$COUNT"; \
	if [ "$$(echo $$COUNT | tr -d ' ')" = "4" ]; then \
		echo "PASS: 4 variants completed (original+medium+small+thumb)"; \
	else \
		echo "FAIL: Expected 4 completed variants"; exit 1; \
	fi

# Upload image to a service and verify worker processed all variants
test-service-image:
	@echo "========================================="
	@echo "Test: Service Image Upload + Worker"
	@echo "========================================="
	@if [ ! -f test-image.jpg ]; then $(MAKE) create-test-image; fi
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Logged in: URL_CODE=$$URL_CODE"; \
	SVC=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Servico Imagem Teste","price":99.90,"duration_minutes":60,"is_active":true}'); \
	SERVICE_ID=$$(echo "$$SVC" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "2. Service created: $$SERVICE_ID"; \
	UPLOAD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/services/$$SERVICE_ID/image \
		-H "Authorization: Bearer $$TOKEN" \
		-F "image=@test-image.jpg;type=image/png"); \
	echo "3. Upload response: $$UPLOAD"; \
	echo "4. Waiting 4s for worker..."; \
	sleep 4; \
	COUNT=$$(docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -t -c \
		"SELECT COUNT(*) FROM images WHERE imageable_id='$$SERVICE_ID' AND processing_status='completed';"); \
	echo "5. Completed variants: $$COUNT"; \
	if [ "$$(echo $$COUNT | tr -d ' ')" = "4" ]; then \
		echo "PASS: 4 variants completed (original+medium+small+thumb)"; \
	else \
		echo "FAIL: Expected 4 completed variants"; exit 1; \
	fi

# Show worker logs and DB summary
test-image-worker-check:
	@echo "===== Worker Logs (last 20 lines) ====="
	@docker compose logs --tail=20 worker-images
	@echo ""
	@echo "===== Images by processing_status ====="
	@docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -c \
		"SELECT processing_status, COUNT(*) FROM images GROUP BY processing_status ORDER BY processing_status;"
	@echo ""
	@echo "===== Pending / Failed images ====="
	@docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -c \
		"SELECT id, imageable_type, imageable_id, variant, processing_status, created_at FROM images WHERE processing_status IN ('pending','failed') ORDER BY created_at DESC LIMIT 20;"

# Run all image tests
test-images-complete: create-test-image test-product-image test-service-image
	@echo ""
	@echo "All image tests passed!"
