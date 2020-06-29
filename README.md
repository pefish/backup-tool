# File Transfer Tool

file_transfer_tool是一款文件传输工具，具有以下特点
1. 支持单个文件以及整个文件夹压缩后的传输
2. 分为服务端以及客户端，支持传输到任意地方
3. 完美支持大文件传输

**注意：本项目主要用于学习研究，成熟的scp完全可以替代本工具**

## Install

```shell script
go get -u github.com/pefish/file-transfer-tool/bin/file_transfer_tool
```

## Quick start

### Server Side

```shell script
file_transfer_tool server --log-level=debug --target-path=~/backup/
```

### Client Side

```shell script
file_transfer_tool client --tcp-address=0.0.0.0:8000 --file=/path/to/file --save-path=test/
```

文件将被传输到服务机上的**~/backup/test/**目录下

## Document

[doc](https://godoc.org/github.com/pefish/backup-tool)

## Security Vulnerabilities

If you discover a security vulnerability, please send an e-mail to [pefish@qq.com](mailto:pefish@qq.com). All security vulnerabilities will be promptly addressed.

## License

This project is licensed under the [Apache License](LICENSE).
