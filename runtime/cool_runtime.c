#include <stdio.h>
#include <stdlib.h>

void print_int(int x) {
    printf("%d\n", x);
}

void print_string(const char* s) {
    printf("%s\n", s);
}

void* coolz_malloc(int size) {
    return malloc(size);
}

void coolz_abort(const char* msg) {
    fprintf(stderr, "ABORT: %s\n", msg);
    exit(1);
}
