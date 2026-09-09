.PHONY: api web test lint build build-lambdas synth

api:
	go run ./cmd/server

web:
	npm --prefix apps/web run dev

test:
	go test -race ./cmd/... ./internal/...
	npm --prefix apps/web run lint
	npm --prefix apps/web run typecheck
	npm --prefix apps/web test
	cd infra && go test ./...

lint:
	go vet ./cmd/... ./internal/...
	npm --prefix apps/web run lint

build:
	go build ./cmd/...
	npm --prefix apps/web run build
	cd infra && go test ./...

build-lambdas:
	mkdir -p build/lambda/api build/lambda/effect-worker build/lambda/outbox-publisher build/lambda/reconciler
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o build/lambda/api/bootstrap ./cmd/lambda/api
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o build/lambda/effect-worker/bootstrap ./cmd/lambda/effect-worker
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o build/lambda/outbox-publisher/bootstrap ./cmd/lambda/outbox-publisher
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o build/lambda/reconciler/bootstrap ./cmd/lambda/reconciler

synth: build-lambdas
	cd infra && npx --yes aws-cdk@2.1131.0 synth --quiet
