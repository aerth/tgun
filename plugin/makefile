export CGO_CFLAGS=-Wall -pedantic
cross_compiler ?= x86_64-w64-mingw32-gcc 
CFLAGS := -Wall -pedantic
goflags := -v -tags netgo,osusergo -ldflags '-w -s'
all: libraries tgun
libraries: tgun.so tgun.a
windows: tgun.dll tgun.dll.a tgun.exe

# binaries
tgun.exe: example_c/example.c tgun.dll.a
	# cross compile win library
	${cross_compiler} -static ${CFLAGS} -o $@ -I. $^
tgun: tgun-static
	mv $^ $@
tgun-shared: example_c/example.c tgun.so
	${CC} ${CFLAGS} -o $@ -I. $^
tgun-static: example_c/example.c tgun.a
	${CC} -static ${CFLAGS} -o $@ -I. $^
tgun-trig-error: example_c/example_error.c tgun.a
	${CC} -static ${CFLAGS} -o $@ -I. $^

# build library
tgun.so: *.go
	go build ${goflags} -buildmode c-shared -o $@
	chmod +x tgun.so
	ldd tgun.so
tgun.a: *.go
	go build ${goflags} -buildmode c-archive -o $@
tgun.dll.a:
	CC=${cross_compiler} CGO_ENABLED=1 GOOS=windows go build ${goflags} -buildmode c-archive -o $@
tgun.dll:
	CC=${cross_compiler} CGO_ENABLED=1 GOOS=windows go build ${goflags} -buildmode c-shared -o $@

clean:
	${RM} tgun*
install:
	mkdir -p /usr/local/include /usr/local/lib /usr/local/bin
	test ! -x tgun || install -v tgun /usr/local/bin/tgun
	install -v tgun.a /usr/local/lib/libtgun.a
	install -v tgun.so /usr/local/lib/libtgun.so
	install -v tgun.h /usr/local/include/tgun.h
	@echo now can use -ltgun to link and '#include <tgun.h>'
