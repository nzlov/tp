# 构建精灵图

## 使用

### 下载可执行文件

### Go

```
go run github.com/nzlov/tp@latest
```

## 使用配置文件

`config.yaml`:

```
output: output
outtype: webp
quality: 90
prefix: tp-
```

use:

```
tp -c config.yaml imagepath...
```

## 使用环境变量

```
PREFIX=ENV tp imagepath...
```

```
PREFIX=ENV tp -c config.yaml imagepath...
```
