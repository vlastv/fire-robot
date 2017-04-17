build:
	mkdir -p bin/
	GOARM=7 GOARCH=arm GOOS=linux go build -ldflags="-w" -o bin/ctl ctl.go fire.go

deploy: build
	scp bin/ctl pi@109.72.66.10:.
