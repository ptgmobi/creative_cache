RM ?= rm -rf
GOBUILD = go build
GOTEST = go test
GOGET = go get -u
APP = bin/ccache

VARS=vars.mk
$(shell ./build_config.sh ${VARS})
include ${VARS}

.PHONY: main deps test clean

main:
	${GOBUILD} -o ${APP} src/main.go

deps:
	${GOGET} github.com/brg-liuwei/gotools

test:
	pushd src/test && ${GOTEST} && popd

clean:
	${RM} bin/* ${VARS}
