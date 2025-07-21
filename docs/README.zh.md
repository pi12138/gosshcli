# gossh

`gossh` 是一个用 Go 编写的命令行工具，它提供了一种简单、安全的方式来管理、连接您的 SSH 服务器并执行命令。

## 功能

- **高级连接管理**: 以交互方式或通过标志添加、列出、删除和分组您的 SSH 连接。
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
4.  **将可执行文件移动到您的 PATH**:
    ```sh
    sudo mv gossh /usr/local/bin/
    ```

## 使用方法

### 管理连接

- **添加新连接 (交互模式)**:
    ```sh
    gossh add -i
    ```
- **添加新连接 (标志模式)**:
    ```sh
    gossh add -n <名称> -u <用户> -H <主机> [-g <分组>] [-p <端口>] [-k <密钥路径>] [-P <密码别名>]
    ```
    - `-i, --interactive`: 使用交互模式。
    - `-g, --group`: 将连接分配到一个分组。

- **列出已保存的连接**:
    ```sh
    gossh list
    gossh list -g <分组名称> # 按分组筛选
    ```

- **列出所有分组**:
    ```sh
    gossh groups
    ```

- **删除连接**:
    ```sh
    gossh remove <连接名称>
    ```

- **测试连接**:
    ```sh
    gossh test <连接名称>
    ```

### 连接与执行

- **连接到服务器 (交互式会话)**:
    ```sh
    gossh connect <连接名称>
    ```

- **执行远程命令**:
    ```sh
    gossh exec <连接名称> "你的命令"
    ```
    示例: `gossh exec web-server "sudo systemctl status nginx"`

### 管理密码

- **添加新密码**:
    ```sh
    gossh password add <别名>
    ```
- **列出已保存的密码别名**:
    ```sh
    gossh password list
    ```
- **删除密码**:
    ```sh
    gossh password remove <别名>
    ```

### 配置管理

- **导出配置**:
    ```sh
    gossh config export > gossh_backup.json
    ```
- **导入配置**:
    ```sh
    gossh config import gossh_backup.json
    ```
    在覆盖您现有配置之前，系统将提示您进行确认。使用 `-f` 或 `--force` 可以跳过提示。

## 配置文件

配置文件存储在 `~/.config/gossh/` 中：

- `config.json`: 存储连接配置。
- `credentials.json`: 存储加密的密码。
- `secret.key`: 用于密码的加密密钥。

**注意**: 不要共享您的 `secret.key` 或 `credentials.json` 文件。

## 贡献

欢迎贡献！请随时提交拉取请求或开启一个 issue。
