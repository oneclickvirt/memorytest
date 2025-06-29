#ifndef MEMORY_H
#define MEMORY_H

#include <stdlib.h>
#include <stdint.h>
#include <string.h>

// 函数声明
double fast_memory_write_test(void* buffer, size_t size);
double fast_memory_read_test(void* buffer, size_t size);
void* aligned_malloc(size_t size, size_t alignment);
long get_page_size(void);

#endif