# Latest Build Size
FROM golang:1.22.6-alpine3.20 AS build-stage

RUN apk update
RUN apk add --no-cache bash git

WORKDIR /autograder-server

COPY . ./full-copy

COPY ./cmd/ cmd/
COPY ./internal/ internal/
COPY ./scripts/ scripts/
COPY ./testdata/ testdata/
COPY ./go.* ./
COPY ./LICENSE ./
COPY ./VERSION.json ./

COPY ./docker/entrypoint.sh ./
RUN chmod 0775 entrypoint.sh

RUN cd full-copy && ./scripts/build.sh 
RUN cd full-copy && AUTOGRADER__DOCKER__DISABLE='true' ./scripts/run_tests.sh
RUN cd full-copy && go run ./cmd/version/main.go --out=VERSION.json && mv VERSION.json ..

RUN rm -rf ./full-copy
RUN apk del git
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

RUN find /go/pkg/ -name 'testdata' | xargs rm -rf
RUN find /go/pkg/ -name '*_test.go' | xargs rm -rf
RUN find /go/pkg/mod/cache/download/ -name '*.zip' | xargs rm 
RUN rm -rf /root/.cache/go-build/


FROM scratch AS final-stage

ENV GOLANG_VERSION=1.22.6
ENV GOTOOLCHAIN=local
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
COPY --from=build-stage / /

WORKDIR /autograder-server

ENTRYPOINT ["./entrypoint.sh"]
