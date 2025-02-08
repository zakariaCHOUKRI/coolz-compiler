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

// Define missing constructors:

void* Int_new() {
    // Just allocate enough memory for an Int object.
    // Adjust size if you store extra fields.
    return malloc(sizeof(void*) + sizeof(int));
}

void* String_new() {
    // Allocate memory for a String object.
    // Change size if needed for additional fields.
    return malloc(sizeof(void*) + sizeof(char*) + sizeof(int));
}

void* Bool_new() {
    // Allocate memory for a Bool object.
    // Update size if needed.
    return malloc(sizeof(void*) + sizeof(char));
}

void* Main_new() {
    // Allocate memory for a Main object.
    // Adjust size for fields if necessary.
    return malloc(sizeof(void*));
}
