# Qssh

轻量快捷的 SSH 终端连接与主机管理工具，基于 Go 语言开发。支持别名快速连接、终端 PTY 交互、远程文件传输及自动备份部署。

---

## 特性

- **快捷直连**：配置主机别名后，直接执行 `qssh <alias>` 即可秒级进入交互终端。
- **自适应登录 Shell**：采用标准 SSH 登录 Shell 调用，原生适配各种系统及默认 Shell（如 OpenWrt/BusyBox 的 `ash`、Linux 的 `bash`、macOS 的 `zsh` 等）。
- **轻量配置**：主机列表统一记录在 `~/.qssh`，简单直观易维护。
- **文件操作与备份**：内置简易的 `copy`（基于 scp 传输）与 `deploy`（自动压缩备份远端旧目录后再覆盖部署）功能。

---

## 编译与安装

确保本地已安装 Go 环境：

```bash
# 编译可执行文件
go build -o qssh main.go

# 推荐将编译产物安装到 PATH 路径（例如 ~/Developer/bin 或 /usr/local/bin）
cp qssh ~/Developer/bin/
# 或
# sudo cp qssh /usr/local/bin/
```

---

## 前置要求

- 认证方式：默认读取本地 `~/.ssh/id_rsa` 私钥进行公钥认证。请确保已将本地公钥配置到远程主机的 `~/.ssh/authorized_keys` 中。
- 配置文件：首次添加主机或运行后会自动创建/读取 `~/.qssh`。

---

## 使用说明

### 1. 命令总览

```bash
qssh <command> [arguments]
```

| 命令 | 参数 | 说明 |
| :--- | :--- | :--- |
| `<alias>` | `<alias>` | 直接以别名连接远程主机交互终端 |
| `connect` | `<alias>` | 连接远程主机交互终端 |
| `add` | `<alias> <user>@<ip>[:<port>]` | 添加新的主机别名配置（默认端口 22） |
| `list` | - | 查看所有已保存的主机别名列表 |
| `copy` | `<alias> <local> [<remote>]` | 复制本地文件/目录到远程主机（默认远端路径为 `~`） |
| `deploy` | `<alias> <local> <remote>` | 自动打包备份远端 `/www/<remote>` 为 zip，并同步本地最新文件 |
| `help` | `-h`, `--help` | 查看帮助提示信息 |

---

### 2. 常见使用示例

#### 添加主机配置
```bash
# 添加端口为 22 的主机
qssh add myserver root@192.168.1.100

# 添加指定自定义端口的主机
qssh add dev root@10.0.0.2:2222
```

#### 连接主机
```bash
# 直接使用别名连接
qssh myserver

# 或者使用 connect 命令
qssh connect myserver
```

#### 查看已配置的主机
```bash
qssh list
```

#### 传输文件
```bash
# 复制本地文件到远端用户家目录 ~
qssh copy myserver ./app.tar.gz

# 复制本地目录到远端指定目录
qssh copy myserver ./build /data/web
```

#### 部署与自动备份
```bash
# 会先在远程将 /www/dist 打包为 dist_YYYY-MM-DD~HH-MM-SS.zip，然后将本地 ./dist 部署至 /www/dist
qssh deploy myserver ./dist dist
```

---

## 配置文件格式 (`~/.qssh`)

配置文件按行记录，每行格式为：`<alias> <user>@<ip>[:<port>]`

示例：
```text
ec20 root@192.168.1.1
myserver root@192.168.1.100:2222
prod admin@47.100.1.10:22
```
