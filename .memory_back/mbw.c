#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/time.h>
#include <time.h>
#include <string.h>
#include <unistd.h>
#include "mbw.h"
#define DEFAULT_NR_LOOPS 10
#define MAX_TESTS 3
#define DEFAULT_BLOCK_SIZE 262144
#define TEST_MEMCPY 0
#define TEST_DUMB 1
#define TEST_MCBLOCK 2
#ifdef _WIN32
#include <windows.h>
int gettimeofday(struct timeval *tv, void *tz)
{
    FILETIME ft;
    unsigned __int64 tmpres = 0;
    if (NULL != tv) {
        GetSystemTimeAsFileTime(&ft);
        tmpres |= ft.dwHighDateTime;
        tmpres <<= 32;
        tmpres |= ft.dwLowDateTime;
        tmpres /= 10;
        tmpres -= 11644473600000000ULL;
        tv->tv_sec = (long)(tmpres / 1000000UL);
        tv->tv_usec = (long)(tmpres % 1000000UL);
    }
    return 0;
}
#endif

long *make_array(unsigned long long asize)
{
    unsigned long long t;
    unsigned int long_size=sizeof(long);
    long *a;
    a=calloc(asize, long_size);
    if(NULL==a) {
        return NULL;
    }
    for(t=0; t<asize; t++) {
        a[t]=0xaa;
    }
    return a;
}

double worker(unsigned long long asize, long *a, long *b, int type, unsigned long long block_size)
{
    unsigned long long t;
    struct timeval starttime, endtime;
    double te;
    unsigned int long_size=sizeof(long);
    unsigned long long array_bytes=asize*long_size;
    if(type==TEST_MEMCPY) {
        gettimeofday(&starttime, NULL);
        memcpy(b, a, array_bytes);
        gettimeofday(&endtime, NULL);
    } else if(type==TEST_MCBLOCK) {
        char* src = (char*)a;
        char* dst = (char*)b;
        gettimeofday(&starttime, NULL);
        for (t=array_bytes; t >= block_size; t-=block_size, src+=block_size){
            dst=(char *) memcpy(dst, src, block_size) + block_size;
        }
        if(t) {
            dst=(char *) memcpy(dst, src, t) + t;
        }
        gettimeofday(&endtime, NULL);
    } else if(type==TEST_DUMB) {
        gettimeofday(&starttime, NULL);
        for(t=0; t<asize; t++) {
            b[t]=a[t];
        }
        gettimeofday(&endtime, NULL);
    }
    te=((double)(endtime.tv_sec*1000000-starttime.tv_sec*1000000+endtime.tv_usec-starttime.tv_usec))/1000000;
    return te;
}

int run_memory_test(unsigned long long mt, struct TestResult *results)
{
    unsigned int long_size=sizeof(long);
    unsigned long long asize=1024*1024/long_size*mt;
    unsigned long long block_size=DEFAULT_BLOCK_SIZE;
    long *a, *b;
    int nr_loops=5;
    int testno, i;
    double te, te_sum;
    if(asize*long_size < block_size) {
        return -1;
    }
    a=make_array(asize);
    if(a==NULL) {
        return -1;
    }
    b=make_array(asize);
    if(b==NULL) {
        free(a);
        return -1;
    }
    for(testno=0; testno<MAX_TESTS; testno++) {
        te_sum=0;
        for (i=0; i<nr_loops; i++) {
            te=worker(asize, a, b, testno, block_size);
            te_sum+=te;
        }
        results[testno].speed = mt/te_sum*nr_loops;
        results[testno].type = testno;
    }
    free(a);
    free(b);
    return 0;
}