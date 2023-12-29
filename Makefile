TARGET_NAME=nas_server

nas_server: clean
	go clean ./...
	go build -o ${TARGET_NAME} ./main.go

clean:
	@if [ -f "${TARGET_NAME}" ]; then \
		rm ${TARGET_NAME}; \
	fi