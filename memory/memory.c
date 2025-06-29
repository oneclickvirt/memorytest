#define _GNU_SOURCE
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <unistd.h>
#include <stdint.h>
#include "memory.h"

// 快速内存写入测试
double fast_memory_write_test(void* buffer, size_t size) {
    struct timeval start, end;
    gettimeofday(&start, NULL);
    // 逐字节写入
    char* ptr = (char*)buffer;
    for (size_t i = 0; i < size; i++) {
        ptr[i] = (char)(i % 256);
    }
    gettimeofday(&end, NULL);
    double elapsed = (end.tv_sec - start.tv_sec) + (end.tv_usec - start.tv_usec) / 1000000.0;
    return elapsed > 0 ? (double)size / elapsed / 1024 / 1024 : 0;
}

// 快速内存读取测试
double fast_memory_read_test(void* buffer, size_t size) {
    struct timeval start, end;
    gettimeofday(&start, NULL);
    char* ptr = (char*)buffer;
    volatile unsigned long long sum = 0;
    for (size_t i = 0; i < size; i++) {
        sum += (unsigned long long)ptr[i];
    }
    gettimeofday(&end, NULL);
    double elapsed = (end.tv_sec - start.tv_sec) + (end.tv_usec - start.tv_usec) / 1000000.0;
    // 防止编译器优化
    if (sum == 0) {
        // 这个分支几乎不会执行
    }
    return elapsed > 0 ? (double)size / elapsed / 1024 / 1024 : 0;
}

// 分配对齐内存
void* aligned_malloc(size_t size, size_t alignment) {
    void* ptr = NULL;
    // 确保alignment是2的幂次方且至少是sizeof(void*)
    if (alignment < sizeof(void*)) {
        alignment = sizeof(void*);
    }
    if (posix_memalign(&ptr, alignment, size) != 0) {
        return NULL;
    }
    return ptr;
}

// 获取系统页面大小
long get_page_size() {
    return sysconf(_SC_PAGESIZE);
}