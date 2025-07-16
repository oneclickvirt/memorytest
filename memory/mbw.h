#ifndef MBW_H
#define MBW_H

#ifdef _WIN32
#include <winsock2.h>
struct timeval {
    long tv_sec;
    long tv_usec;
};

struct timezone {
    int tz_minuteswest;
    int tz_dsttime;
};

int gettimeofday(struct timeval *tv, struct timezone *tz);
#endif

struct TestResult {
    int type;
    double speed;
};

int run_memory_test(unsigned long long mt, struct TestResult *results);

#endif