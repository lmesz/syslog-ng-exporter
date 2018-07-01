build:
	GOOS=linux GOARCH=amd64 go build
copy:
	docker cp read_from_socket infra_web_1:/tmp
