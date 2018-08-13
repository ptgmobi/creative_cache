RM ?= rm -rf
GOBUILD = go build
GOTEST = go test
GOGET = go get -u
APP = bin/ccache
GUARD = bin/cache_guard

VARS=vars.mk
$(shell ./build_config.sh ${VARS})
include ${VARS}

.PHONY: main deps test clean

main:
	${GOBUILD} -o ${APP} src/main.go
	${GOBUILD} -o ${GUARD} src/ccache_guard.go

deps:
	${GOGET} github.com/brg-liuwei/gotools
	${GOGET} github.com/garyburd/redigo/redis

test:
	pushd src/loop && ${GOTEST} && popd

clean:
	${RM} bin/* ${VARS}
