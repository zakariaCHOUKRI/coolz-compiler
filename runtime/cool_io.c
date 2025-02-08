#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// We assume "String" is stored as [vtablePtr, char*, int length]
// We assume "Int"    is stored as [vtablePtr, int value]

void* IO_out_string(void* self, void* x) {
    char* str = *((char**)((char*)x + sizeof(void*)));
    printf("%s", str);
    return self;
}

void* IO_out_int(void* self, void* x) {
    int value = *((int*)((char*)x + sizeof(void*)));
    printf("%d", value);
    return self;
}

void* IO_in_string(void* self) {
    char buffer[1024];
    // Read up to a newline
    if (!fgets(buffer, sizeof(buffer), stdin)) {
        // Handle EOF or error
        return NULL;
    }
    // Strip trailing newline
    size_t len = strlen(buffer);
    if (len > 0 && buffer[len - 1] == '\n') {
        buffer[len - 1] = '\0';
        len--;
    }
    // Allocate a new String object
    void* strObj = String_new();
    // Copy contents
    char** strField = (char**)((char*)strObj + sizeof(void*));
    *strField = malloc(len + 1);
    strcpy(*strField, buffer);
    // Set length field
    int* lenField = (int*)((char*)strObj + sizeof(void*) + sizeof(char*));
    *lenField = (int)len;
    return strObj;
}

void* IO_in_int(void* self) {
    int value = 0;
    // Read an integer from stdin, ignoring non-digit prefix
    if (scanf("%d", &value) != 1) {
        // Handle invalid input
        return NULL;
    }
    // Create new Int object
    void* intObj = Int_new();
    int* intField = (int*)((char*)intObj + sizeof(void*));
    *intField = value;
    return intObj;
}