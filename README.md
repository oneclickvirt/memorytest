# memoryTest

[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2FmemoryTest&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://hits.seeyoufarm.com) [![Build and Release](https://github.com/oneclickvirt/memoryTest/actions/workflows/main.yml/badge.svg)](https://github.com/oneclickvirt/memoryTest/actions/workflows/main.yml)

内存测试模块 (Memory Test Module) 

# 功能(Features)

- [x] 使用```sysbench```测试内存的顺序读写IO
- [x] 使用```dd```测试内存的读写IO
- [x] 使用```winsat```测试内存的读写性能
- [x] 以```-l```指定输出的语言类型，可指定```zh```或```en```，默认不指定时使用中文输出
- [x] 以```-m```指定测试的方法，可指定```sysbench```或```dd```，默认不指定时使用```sysbench```进行测试
- [x] 全平台编译支持

注意：默认不自动安装```sysbench```组件，如需使用请自行安装后再使用本项目，如```apt update && apt install sysbench -y```

# 使用(Usage)

```
curl https://raw.githubusercontent.com/oneclickvirt/memoryTest/main/mt_install.sh -sSf | sh
```
