all: getdeps install

checkdeps:
	@./checkdeps.sh

getdeps: checkdeps
	@go get github.com/tools/godep && echo "Installed godep"
	@go get golang.org/x/tools/cmd/cover && echo "Installed cover"

install:
	@godep go install github.com/minio-io/mc/new-cmd && echo "Installed new-cmd"

s3:
	@godep go test -race -coverprofile=cover.out github.com/minio-io/mc/pkg/s3
