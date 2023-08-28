#include <cstdio>
#include <tgun.h>
#include <iostream>

#include <stdlib.h>
#include <stdarg.h>
#include <thread>
#include <chrono>
#define xaddr "127.0.0.1"
#define xport 8080
#define XBUFSIZE 1024

// easy sprintf allocator from man page
char *make_message(const char *fmt, ...){
    int n = 0;
    size_t size = 0;
    char *p = NULL;
    va_list ap;
    va_start(ap, fmt);
    n = vsnprintf(p, size, fmt, ap);
    va_end(ap);
    if (n < 0)
        return NULL;
    size = (size_t) n + 1;
    p = (char*) malloc(size);
    if (p == NULL)
        return NULL;
    va_start(ap, fmt);
    n = vsnprintf(p, size, fmt, ap);
    va_end(ap);
    if (n < 0) {
        free(p);
        return NULL;
    }
    return p;
}
using namespace std::chrono_literals;
int main(){
    printf("hello tgun v%s\n", tgun_version());
    bool useSSL = true;
    bool useUnsafeSSL = true;
    int fd = tgun_connect(xaddr, xport, useSSL, useUnsafeSSL); // ssl=true, unsafe=true
    if (fd < 0) {
        printf("error connecting: %d -- %s\n", fd, tgunerr());
        return 1;
    }
    printf("Open TCP Connection Num. %d\n", fd);
    size_t max_size = XBUFSIZE;
    char buf[XBUFSIZE+1];
    for (int attempt = 1; attempt <= 3; attempt++){
        int n = tgun_read(fd, buf, max_size);
        if (n < 0) {
            printf("error reading! %d -- %s\n", n, tgunerr());
            return 1;
        }
        buf[n] = '\0';
        printf("read %d bytes: %s\n", n, buf);
        char* out =make_message("it works! %d/3", attempt);
        n = tgun_write(fd, out);
        free(out);
        printf("wrote %d bytes\n", n);
        std::this_thread::sleep_for(1000ms);
    }

    char* out =make_message("/quit");
    tgun_write(fd, out);
    free(out);

    std::this_thread::sleep_for(2000ms);
    tgun_disconnect(fd);
    printf("disconnected: %d\n", fd);
    fflush(stdout);
    return 0;
}
