export CGO_CFLAGS=-Wall -pedantic
CFLAGS := -Wall -pedantic
goflags := -v -tags netgo,osusergo -ldflags '-w -s'
all: tgun.so tgun.a tgun
tgun: example_c/example.c tgun.a
	${CC} -static ${CFLAGS} -o $@ -I. $^
tgun.so: *.go
	go build ${goflags} -buildmode c-shared -o $@
	chmod +x tgun.so
	ldd tgun.so
tgun.a: *.go
	go build ${goflags} -buildmode c-archive -o $@
a.out: example_c/example_error.c tgun.a
	${CC} ${CFLAGS} -o $@ -I. $^
clean:
	${RM} tgun.a tgun.h tgun.so tgun
install:
	install tgun.a /usr/local/lib/libtgun.a
	install tgun.so /usr/local/lib/libtgun.so
	mkdir -p /usr/local/include
	install tgun.h /usr/local/include/tgun.h
	@echo now can use -ltgun to link and '#include <tgun.h>'
