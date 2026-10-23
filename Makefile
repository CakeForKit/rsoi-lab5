.PHONY: test test-go test-api build push deploy lint

MODULES := lb-core ms-flight ms-ticket ms-bonus ms-gateway
NAMESPACE ?= rsoi
IMAGE_TAG ?= latest
DOCKER_REGISTRY ?= local

test-api:
	docker compose up -d
	WAIT_PORTS=8080,8070,8060,8050 ./scripts/wait-script.sh
	newman run --delay-request=100 --folder='Gateway API' -e v1/postman/environment.json v1/postman/collection.json


