#ifndef MBW_H
#define MBW_H

#ifdef _WIN32
#include <winsock2.h>
#include <time.h>
#include <sys/time.h>
#ifndef _TIMEVAL_DEFINED
#define _TIMEVAL_DEFINED
struct timeval {
    long tv_sec;
    long tv_usec;
};
#endif
int gettimeofday(struct timeval *tv, void *tz);
#endif

struct TestResult {
    int type;
    double speed;
};

int run_memory_test(unsigned long long mt, struct TestResult *results);

#endif
