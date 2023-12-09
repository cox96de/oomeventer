FROM hub.byted.org/codebase/ci_go_1_20 as builder
RUN apt-get update && apt-get install -y llvm clang
ADD . /src
WORKDIR /src
RUN ./build.sh
FROM hub.byted.org/toutiao.debian
COPY --from=builder /src/oomeventer /usr/local/bin/
ENTRYPOINT /usr/local/bin/oomeventer