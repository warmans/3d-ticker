build: build.linux build.windows

.PHONY: build.linux
build.linux:
	go build -o build/3dticker-lin

.PHONY: build.windows
build.windows:
	GOOS=windows GOARCH=386 go build -ldflags -H=windowsgui -o build/3dticker-win.exe

.PHONY: build.windows.debug
build.windows.debug:
	GOOS=windows GOARCH=386 go build -o build/3dticker-win.exe
