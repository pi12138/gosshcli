# gossh 使用指南

`gossh` 是一个用 Go 编写的命令行工具，它提供了一种简单、安全的方式来管理、连接您的 SSH 服务器并执行命令。

## 功能

- **连接管理**: 以交互方式或通过命令行标志添加、列出、删除和分组您的 SSH 连接。
- **远程命令执行**: 在服务器上执行命令，而无需启动完整的交互式会话。
- **安全密码存储**: 密码使用 AES-256 加密并存储在本地。
- **连接测试**: 无需完全登录即可测试到服务器的连接和身份验证。
- **配置可移植性**: 将您的所有连接和凭证导出到单个文件以进行备份或迁移，并在另一台机器上导入它们。
- **多种身份验证方法**: 支持密码、私钥（包括带密码的私钥）和交互式密码身份验证。

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
4.  **将可执行文件移动到您的 PATH 中**:
    ```sh
    sudo mv gossh /usr/local/bin/
    ```

## 命令用法

### 管理连接 (`config` 命令)

使用 `gossh config` 子命令来管理您的连接配置。

- **添加新连接**:
    ```sh
    gossh config add [flags]
    ```
    **模式**:
    - **交互模式**: `gossh config add -i` 或 `gossh config add --interactive`
    - **标志模式**:
        ```sh
        gossh config add -n <名称> -u <用户> -H <主机> [其他标志...]
        ```
    **标志**:
    - `-i, --interactive`: 使用交互式会谈来添加新连接。
    - `-n, --name`: 连接名称 (在非交互模式下为必需)。
    - `-u, --user`: 用户名 (在非交互模式下为必需)。
    - `-H, --host`: 主机地址 (在非交互模式下为必需)。
    - `-g, --group`: 将连接分配到一个分组。
    - `-p, --port`: 端口号 (默认为 22)。
    - `-k, --key`: 私钥的路径。
    - `-P, --use-password`: 使用已保存的密码别名进行身份验证。

- **列出已保存的连接**:
    ```sh
    gossh config list [flags]
    ```
    **标志**:
    - `-g, --group`: 按分组名称筛选连接。

- **删除连接**:
    ```sh
    gossh config remove <连接名称>
    ```

### 连接与执行

- **连接到服务器 (交互式会话)**:
    ```sh
    gossh connect <连接名称>
    ```

- **执行远程命令**:
    ```sh
    gossh exec <连接名称> "<你的命令>"
    ```
    示例: `gossh exec web-server "sudo systemctl status nginx"`

- **复制文件 (上传/下载)**:
    ```sh
    gossh scp [源路径] [目标路径]
    ```
    **上传** (本地到远程):
    ```sh
    gossh scp /本地/文件.txt 服务器:/远程/路径/
    gossh scp -r /本地/目录 服务器:/远程/路径/
    ```
    **下载** (远程到本地):
    ```sh
    gossh scp 服务器:/远程/文件.txt /本地/路径/
    gossh scp -r 服务器:/远程/目录 /本地/路径/
    ```
    **标志**:
    - `-r, --recursive`: 递归复制整个目录。

- **测试连接**:
    ```sh
    gossh test <连接名称>
    ```
    此命令会尝试进行身份验证然后立即断开连接，以验证配置是否正确。

### 管理密码 (`password` 命令)

安全地存储和管理用于连接的密码。

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

### 管理分组

- **列出所有分组**:
    ```sh
    gossh groups
    ```

### 导入与导出

- **导出配置**:
    ```sh
    gossh config export > gossh_backup.json
    ```
    此命令会将所有连接和加密的凭证打印到标准输出。

- **导入配置**:
    ```sh
    gossh config import <文件路径>
    ```
    从 `export` 命令创建的文件中导入配置。
    **警告**: 这将覆盖您现有的所有配置。
    **标志**:
    - `-f, --force`: 跳过覆盖前的确认提示。

## 配置文件

配置文件存储在 `~/.config/gossh/` 目录中：

- `config.json`: 存储连接配置。
- `credentials.json`: 存储加密的密码。
- `secret.key`: 用于密码加密的密钥。

**注意**: 请勿共享您的 `secret.key` 或 `credentials.json` 文件，因为它们包含敏感信息。

---
本项目由 pi12138 发起并引导开发，Qwen 和 Google Gemini 参与了开发过程。