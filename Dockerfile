FROM golang:1.24 as builder
RUN apt-get update && apt-get install -y llvm clang
ADD . /src
WORKDIR /src
# Make it build faster in China mainland.
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN ./build.sh
FROM debian
COPY --from=builder /src/oomeventer /usr/local/bin/
ENTRYPOINT /usr/local/bin/oomeventer