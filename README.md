# memoryTest

[![Hits](https://hits.spiritlhl.net/memorytest.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false)](https://hits.spiritlhl.net)

[![Build and Release](https://github.com/oneclickvirt/memorytest/actions/workflows/build.yaml/badge.svg)](https://github.com/oneclickvirt/memorytest/actions/workflows/build.yaml)

内存测试模块 (Memory Test Module) 

# 功能(Features)

- [x] 支持使用```stream```进行高性能内存带宽测试 (最高优先级，无需root权限)
- [x] 支持使用```dd```测试内存的读写IO (需要root权限)
- [x] 支持使用```sysbench```测试内存的顺序读写IO (需要root权限)
- [x] 支持使用```winsat```测试内存的读写性能
- [x] 支持Go自身静态依赖注入[dd](https://github.com/oneclickvirt/dd)，使用时无额外环境依赖需求
- [x] 以```-l```指定输出的语言类型，可指定```zh```或```en```，默认不指定时使用中文输出
- [x] 以```-m```指定测试的方法，可指定```stream```或```dd```或```sysbench```或```winsat```或```auto```，默认不指定时按优先级自动选择测试方法
- [x] 全平台编译支持，支持无权限测试时优先尝试STREAM，若不可用再使用C重构或自编译的mbw程序模拟大内存COPY块测试内存性能

## 测试方法优先级
当不指定`-m`参数时，程序按以下优先级自动选择测试方法：
1. **STREAM** - 如果检测到stream二进制文件，优先使用(无需root权限)
2. **DD** - 如果STREAM不可用，使用DD测试(需要root权限)
3. **Sysbench** - 作为最终的兜底实现(需要root权限)

**无root权限时的特殊处理：**
- 当检测到系统无root权限时，DD和Sysbench测试会自动优先尝试STREAM方式
- 如果STREAM不可用，则回退使用mbw程序进行内存测试

注意：默认不自动安装```sysbench```组件，如需使用请自行安装后再使用本项目，如```apt update && apt install sysbench -y```

# 使用(Usage)

下载及安装

```
curl https://raw.githubusercontent.com/oneclickvirt/memoryTest/main/mt_install.sh -sSf | bash
```

使用

```
memorytest
```

或

```
./memorytest
```

进行测试

```
Usage: memorytest [options]
  -h    Show help information
  -l string
        Language parameter (en or zh)
  -log
        Enable logging
  -m string
        Specific Test Method (stream or dd or sysbench or winsat)
  -v    show version
```

更多架构请查看 https://github.com/oneclickvirt/memorytest/releases/tag/output

## 卸载

```
rm -rf /root/memorytest
rm -rf /usr/bin/memorytest
```

## 在Golang中使用

```
go get github.com/oneclickvirt/memorytest@v0.0.9-20250924135209
```

## 测试图

sysbench测试

![图片](https://github.com/oneclickvirt/memoryTest/assets/103393591/741689a2-7887-4cec-9df5-c8e309b2dd84)

dd测试

![图片](https://github.com/oneclickvirt/memoryTest/assets/103393591/34de9add-dbf6-44dd-91cc-b7102de66d3f)

winsat测试

![1716466171182](https://github.com/oneclickvirt/memoryTest/assets/103393591/c8d38d4e-7357-4c27-b55b-4703805a5cb9)

mbw测试

![f4cc8695e41070f9c393071c49315464](https://github.com/user-attachments/assets/10538fb0-3d4e-4118-b248-8ccfd6a09e24)


