# memoryTest

[![Hits](https://hits.spiritlhl.net/memorytest.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false)](https://hits.spiritlhl.net)

[![Build and Release](https://github.com/oneclickvirt/memoryTest/actions/workflows/main.yml/badge.svg)](https://github.com/oneclickvirt/memoryTest/actions/workflows/main.yml)

内存测试模块 (Memory Test Module) 

# 功能(Features)

- [x] 支持使用```sysbench```测试内存的顺序读写IO
- [x] 支持使用```dd```测试内存的读写IO
- [x] 支持使用```winsat```测试内存的读写性能
- [x] 支持Go自身静态依赖注入[dd](https://github.com/oneclickvirt/dd)，使用时无额外环境依赖需求
- [x] 以```-l```指定输出的语言类型，可指定```zh```或```en```，默认不指定时使用中文输出
- [x] 以```-m```指定测试的方法，可指定```sysbench```或```dd```，默认不指定时使用```sysbench```进行测试
- [x] 全平台编译支持

注意：默认不自动安装```sysbench```组件，如需使用请自行安装后再使用本项目，如```apt update && apt install sysbench -y```

## TODO

- [ ] 正式测试前检测当前路径挂载盘剩余空间是否足够生成测试文件
- [ ] 优化测试失败时的报错和输出
- [ ] 修复WIN系统的虚拟下的CPU测试无法使用winsat的问题
- [ ] 适配MACOS系统测试

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
        Specific Test Method (sysbench or dd)
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
go get github.com/oneclickvirt/memorytest@v0.0.5-20250404130028
```

## 测试图

sysbench测试

![图片](https://github.com/oneclickvirt/memoryTest/assets/103393591/741689a2-7887-4cec-9df5-c8e309b2dd84)

dd测试

![图片](https://github.com/oneclickvirt/memoryTest/assets/103393591/34de9add-dbf6-44dd-91cc-b7102de66d3f)

winsat测试

![1716466171182](https://github.com/oneclickvirt/memoryTest/assets/103393591/c8d38d4e-7357-4c27-b55b-4703805a5cb9)

