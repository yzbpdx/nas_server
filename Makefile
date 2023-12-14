TARGET_NAME=nas_server

nas_server: clean
	go clean ./...
	go build -o ${TARGET_NAME} ./main.go

clean:
	rm ${TARGET_NAME}