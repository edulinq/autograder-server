FROM golang:1.22.6-alpine3.20 AS build-stage

RUN apk update
RUN apk add --no-cache bash git

WORKDIR /autograder-server

# Extract version information from the Git repository and save it to VERSION.json.
COPY . ./full-copy
RUN cd full-copy && go run cmd/version/main.go --out=../VERSION.json
RUN rm -rf ./full-copy

# Copie over only what's nessesary.
COPY ./cmd/ cmd/
COPY ./internal/ internal/
COPY ./scripts/ scripts/
COPY ./testdata/ testdata/
COPY ./go.* ./
COPY ./LICENSE ./

COPY ./docker/entrypoint.sh ./
RUN chmod 0775 entrypoint.sh

# Build and test to verify the binaries will always run.
RUN ./scripts/build.sh
RUN AUTOGRADER__DOCKER__DISABLE='true' ./scripts/run_tests.sh
RUN rm -rf ./bin

# Clean up unnecessary files and directories to reduce the image size.
# - Uninstall unesseray packages.
RUN apk del git

# - Remove 'testdata' directories from Go packages, these are not nessesary in production.
# - Delete test files (*_test.go) from Go packages, these are not nessesary in production.
# - Remove cached ZIP files from Go module downloads, these are not nessesary for building only for downloading.
RUN find /go/pkg/ -name 'testdata' | xargs rm -rf
RUN find /go/pkg/ -name '*_test.go' | xargs rm -rf
RUN find /go/pkg/mod/cache/download/ -name '*.zip' | xargs rm

# - Clear the Go build cache.
RUN rm -rf /root/.cache/go-build/

# - Remove unnessesary tools in Go sorce code and packages.
# These are need for debugging and testing which we dont need in Autograder image.
RUN rm -rf \
    /usr/local/go/api \
    /usr/local/go/test \
    /usr/local/go/pkg/tool/linux_arm64/pack \
    /usr/local/go/pkg/tool/linux_arm64/buildid \
    /usr/local/go/pkg/tool/linux_arm64/nm \
    /usr/local/go/pkg/tool/linux_arm64/addr2line \
    /usr/local/go/pkg/tool/linux_arm64/covdata \
    /usr/local/go/pkg/tool/linux_arm64/fix \
    /usr/local/go/pkg/tool/linux_arm64/doc \
    /usr/local/go/pkg/tool/linux_arm64/objdump \
    /usr/local/go/pkg/tool/linux_arm64/cgo \
    /usr/local/go/pkg/tool/linux_arm64/pprof \
    /usr/local/go/pkg/tool/linux_arm64/trace \
    /usr/local/go/src/cmd/pack \
    /usr/local/go/src/cmd/buildid \
    /usr/local/go/src/cmd/nm \
    /usr/local/go/src/cmd/addr2line \
    /usr/local/go/src/cmd/covdata \
    /usr/local/go/src/cmd/fix \
    /usr/local/go/src/cmd/doc \
    /usr/local/go/src/cmd/objdump \
    /usr/local/go/src/cmd/cgo \
    /usr/local/go/src/cmd/pprof \
    /usr/local/go/src/cmd/trace

FROM scratch AS final-stage

# Since we're starting from a empty base image (scratch) we need to set the environment variables.
ENV GOLANG_VERSION=1.22.6
ENV GOTOOLCHAIN=local
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
COPY --from=build-stage / /

WORKDIR /autograder-server

ENTRYPOINT ["./entrypoint.sh"]
