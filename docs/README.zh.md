# gossh

`gossh` 是一个用 Go 编写的命令行工具，它提供了一种简单而安全的方式来管理和连接到您的 SSH 服务器。它允许您保存连接配置、管理加密的密码，并使用简单的命令轻松连接到您的服务器。

## 功能

- **连接管理**: 添加、列出和删除 SSH 连接配置。
- **安全密码存储**: 密码使用 AES-256 加密并存储在本地。
- **多种身份验证方法**: 支持密码、私钥和交互式密码身份验证。
- **易于使用**: 简单直观的命令行界面。

## 安装

1.  **先决条件**: 确保您的系统上已安装 Go。
2.  **克隆仓库**:
    ```sh
    git clone https://github.com/your-username/gossh.git
    cd gossh
    ```
3.  **构建可执行文件**:
    ```sh
    go build -o gossh .
    ```
4.  **将可执行文件移动到您的 PATH**:
    ```sh
    sudo mv gossh /usr/local/bin/
    ```

## 使用方法

### 管理连接

- **添加新连接**:
    ```sh
    gossh add -n <连接名称> -u <用户名> -H <主机> [-p <端口>] [-k <密钥路径>] [-P <密码别名>]
    ```
    - `-n`: 连接名称 (必需)
    - `-u`: 用户名 (必需)
    - `-H`: 主机地址 (必需)
    - `-p`: 端口号 (默认为 22)
    - `-k`: 私钥路径
    - `-P`: 按别名使用已保存的密码

- **列出已保存的连接**:
    ```sh
    gossh list
    ```

- **删除连接**:
    ```sh
    gossh remove <连接名称>
    ```

### 管理密码

- **添加新密码**:
    ```sh
    gossh password add <别名>
    ```
    系统将提示您输入密码。

- **列出已保存的密码别名**:
    ```sh
    gossh password list
    ```

- **删除密码**:
    ```sh
    gossh password remove <别名>
    ```

### 连接到服务器

- **连接到已保存的连接**:
    ```sh
    gossh connect <连接名称>
    ```

## 配置

配置文件存储在 `~/.config/gossh/` 中：

- `config.json`: 存储连接配置。
- `credentials.json`: 存储加密的密码。
- `secret.key`: 用于密码的加密密钥。

**注意**: 不要共享您的 `secret.key` 或 `credentials.json` 文件。

## 贡献

欢迎贡献！请随时提交拉取请求或开启一个 issue。
