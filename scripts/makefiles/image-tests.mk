.PHONY: create-test-image test-product-image test-service-image \
        test-image-worker-check test-images-complete test-multi-image \
        test-image-title test-image-delete test-image-list

# Creates a valid 10x10 PNG test file at test-image.jpg using Python
create-test-image:
	@python3 scripts/gen_test_image.py test-image.jpg

# Upload image to a product and verify worker processed (single row model)
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
	UPLOAD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "images=@test-image.jpg;type=image/png"); \
	echo "3. Upload response: $$UPLOAD"; \
	echo "4. Waiting 5s for worker..."; \
	sleep 5; \
	COUNT=$$(docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -t -c \
		"SELECT COUNT(*) FROM images WHERE imageable_id='$$PRODUCT_ID' AND processing_status='completed';"); \
	echo "5. Completed images: $$COUNT"; \
	if [ "$$(echo $$COUNT | tr -d ' ')" = "1" ]; then \
		echo "PASS: 1 image row completed (all variants in single row)"; \
	else \
		echo "FAIL: Expected 1 completed image row"; exit 1; \
	fi; \
	echo "6. Checking variant URLs filled..."; \
	VARIANTS=$$(docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -t -c \
		"SELECT CASE WHEN medium_url IS NOT NULL AND small_url IS NOT NULL AND thumb_url IS NOT NULL THEN 'ok' ELSE 'missing' END FROM images WHERE imageable_id='$$PRODUCT_ID' AND processing_status='completed';"); \
	if [ "$$(echo $$VARIANTS | tr -d ' ')" = "ok" ]; then \
		echo "PASS: All variant URLs populated"; \
	else \
		echo "FAIL: Some variant URLs missing"; exit 1; \
	fi

# Upload image to a service and verify worker processed (single row model)
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
	UPLOAD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/services/$$SERVICE_ID/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "images=@test-image.jpg;type=image/png"); \
	echo "3. Upload response: $$UPLOAD"; \
	echo "4. Waiting 5s for worker..."; \
	sleep 5; \
	COUNT=$$(docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -t -c \
		"SELECT COUNT(*) FROM images WHERE imageable_id='$$SERVICE_ID' AND processing_status='completed';"); \
	echo "5. Completed images: $$COUNT"; \
	if [ "$$(echo $$COUNT | tr -d ' ')" = "1" ]; then \
		echo "PASS: 1 image row completed (all variants in single row)"; \
	else \
		echo "FAIL: Expected 1 completed image row"; exit 1; \
	fi

# Test multi-image upload
test-multi-image:
	@echo "========================================="
	@echo "Test: Multi-Image Upload"
	@echo "========================================="
	@if [ ! -f test-image.jpg ]; then $(MAKE) create-test-image; fi
	@python3 scripts/gen_test_image.py test-image2.jpg
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Logged in: URL_CODE=$$URL_CODE"; \
	PROD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Produto Multi Imagem","price":29.90,"stock":5,"is_active":true}'); \
	PRODUCT_ID=$$(echo "$$PROD" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "2. Product created: $$PRODUCT_ID"; \
	UPLOAD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "images=@test-image.jpg;type=image/png" \
		-F "images=@test-image2.jpg;type=image/png"); \
	echo "3. Upload response: $$UPLOAD"; \
	IMG_COUNT=$$(echo "$$UPLOAD" | grep -o '"id"' | wc -l); \
	echo "4. Images uploaded: $$IMG_COUNT"; \
	if [ "$$(echo $$IMG_COUNT | tr -d ' ')" = "2" ]; then \
		echo "PASS: 2 images uploaded"; \
	else \
		echo "FAIL: Expected 2 images uploaded"; exit 1; \
	fi

# Test image title update
test-image-title:
	@echo "========================================="
	@echo "Test: Update Image Title"
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
		-d '{"name":"Produto Titulo Imagem","price":19.90,"stock":3,"is_active":true}'); \
	PRODUCT_ID=$$(echo "$$PROD" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "2. Product created: $$PRODUCT_ID"; \
	UPLOAD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "images=@test-image.jpg;type=image/png"); \
	IMAGE_ID=$$(echo "$$UPLOAD" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "3. Image uploaded: $$IMAGE_ID"; \
	UPDATE=$$(curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/images/$$IMAGE_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"title":"Foto do Produto","alt_text":"Imagem principal","translations":{"title":{"en":"Product Photo","es":"Foto del Producto"}}}'); \
	echo "4. Update response: $$UPDATE"; \
	echo "$$UPDATE" | grep -q "image_updated" && echo "PASS: Image title updated" || (echo "FAIL: Title update failed"; exit 1)

# Test image listing
test-image-list:
	@echo "========================================="
	@echo "Test: List Product Images"
	@echo "========================================="
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Logged in: URL_CODE=$$URL_CODE"; \
	FIRST_PRODUCT=$$(docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -t -c \
		"SELECT id FROM products WHERE tenant_id=(SELECT id FROM tenants WHERE url_code='$$URL_CODE') LIMIT 1;" | tr -d ' '); \
	echo "2. Product: $$FIRST_PRODUCT"; \
	LIST=$$(curl -s http://localhost:8080/api/v1/$$URL_CODE/products/$$FIRST_PRODUCT/images \
		-H "Authorization: Bearer $$TOKEN"); \
	echo "3. Images list: $$LIST"; \
	echo "$$LIST" | grep -q "images" && echo "PASS: Image list returned" || echo "INFO: No images found (or endpoint issue)"

# Test image deletion
test-image-delete:
	@echo "========================================="
	@echo "Test: Delete Image"
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
		-d '{"name":"Produto Delete Img","price":9.90,"stock":1,"is_active":true}'); \
	PRODUCT_ID=$$(echo "$$PROD" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "2. Product created: $$PRODUCT_ID"; \
	UPLOAD=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "images=@test-image.jpg;type=image/png"); \
	IMAGE_ID=$$(echo "$$UPLOAD" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "3. Image uploaded: $$IMAGE_ID"; \
	DELETE=$$(curl -s -X DELETE http://localhost:8080/api/v1/$$URL_CODE/images/$$IMAGE_ID \
		-H "Authorization: Bearer $$TOKEN"); \
	echo "4. Delete response: $$DELETE"; \
	echo "$$DELETE" | grep -q "image_deleted" && echo "PASS: Image deleted" || (echo "FAIL: Delete failed"; exit 1)

# Show worker logs and DB summary
test-image-worker-check:
	@echo "===== Worker Logs (last 20 lines) ====="
	@docker compose logs --tail=20 worker-images
	@echo ""
	@echo "===== Images by processing_status ====="
	@docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -c \
		"SELECT processing_status, COUNT(*) FROM images GROUP BY processing_status ORDER BY processing_status;"
	@echo ""
	@echo "===== Images detail ====="
	@docker exec saas-single-db-api-postgres-1 psql -U saasuser -d saasdb -c \
		"SELECT id, imageable_type, processing_status, original_path, medium_url IS NOT NULL AS has_medium, small_url IS NOT NULL AS has_small, thumb_url IS NOT NULL AS has_thumb, created_at FROM images ORDER BY created_at DESC LIMIT 20;"

# Run all image tests
test-images-complete: create-test-image test-product-image test-service-image test-multi-image test-image-title test-image-delete
	@echo ""
	@echo "All image tests passed!"
