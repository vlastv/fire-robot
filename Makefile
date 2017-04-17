build:
	mkdir -p bin/
	GOARM=7 GOARCH=arm GOOS=linux go build -ldflags="-w" -o bin/fire main.go fire.go

deploy: build
	scp bin/fire pi@109.72.66.10:.
