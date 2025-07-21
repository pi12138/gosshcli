# gossh

[简体中文](docs/README.zh.md)

`gossh` is a command-line tool written in Go that provides a simple and secure way to manage, connect, and execute commands on your SSH servers.

## Features

- **Advanced Connection Management**: Add (interactively or via flags), list, remove, and group your SSH connections.
- **Remote Command Execution**: Execute commands on a server without starting a full interactive session.
- **Secure Password Storage**: Passwords are encrypted using AES-256 and stored locally.
- **Connection Testing**: Test connectivity and authentication to a server without a full login.
- **Configuration Portability**: Export all your connections and credentials to a single file for backup or migration, and import them on another machine.
- **Multiple Authentication Methods**: Supports password, private key (with passphrase), and interactive password authentication.

## Installation

1.  **Prerequisites**: Ensure you have Go installed on your system.
2.  **Clone the repository**:
    ```sh
    git clone https://github.com/your-username/gossh.git
    cd gossh
    ```
3.  **Build the executable**:
    ```sh
    go build -o gossh .
    ```
4.  **Move the executable to your PATH**:
    ```sh
    sudo mv gossh /usr/local/bin/
    ```

## Usage

### Managing Connections

- **Add a new connection (Interactive)**:
    ```sh
    gossh add -i
    ```
- **Add a new connection (Flags)**:
    ```sh
    gossh add -n <name> -u <user> -H <host> [-g <group>] [-p <port>] [-k <key-path>] [-P <password-alias>]
    ```
    - `-i, --interactive`: Use interactive mode.
    - `-g, --group`: Assign the connection to a group.

- **List saved connections**:
    ```sh
    gossh list
    gossh list -g <group-name> # Filter by group
    ```

- **List all groups**:
    ```sh
    gossh groups
    ```

- **Remove a connection**:
    ```sh
    gossh remove <connection-name>
    ```

- **Test a connection**:
    ```sh
    gossh test <connection-name>
    ```

### Connecting and Executing

- **Connect to a server (Interactive Session)**:
    ```sh
    gossh connect <connection-name>
    ```

- **Execute a remote command**:
    ```sh
    gossh exec <connection-name> "your command here"
    ```
    Example: `gossh exec web-server "sudo systemctl status nginx"`

### Managing Passwords

- **Add a new password**:
    ```sh
    gossh password add <alias>
    ```
- **List saved password aliases**:
    ```sh
    gossh password list
    ```
- **Remove a password**:
    ```sh
    gossh password remove <alias>
    ```

### Configuration Management

- **Export configuration**:
    ```sh
    gossh config export > gossh_backup.json
    ```
- **Import configuration**:
    ```sh
    gossh config import gossh_backup.json
    ```
    You will be prompted for confirmation before your existing configuration is overwritten. Use `-f` or `--force` to skip the prompt.

## Configuration Files

Configuration files are stored in `~/.config/gossh/`:

- `config.json`: Stores connection configurations.
- `credentials.json`: Stores encrypted passwords.
- `secret.key`: The encryption key for your passwords.

**Note**: Do not share your `secret.key` or `credentials.json` files.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.
