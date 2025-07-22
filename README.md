# gossh

[简体中文](docs/README.zh.md)

`gossh` is a command-line tool written in Go that provides a simple and secure way to manage, connect, and execute commands on your SSH servers.

## Features

- **Connection Management**: Add (interactively or via flags), list, remove, and group your SSH connections.
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

### Managing Connections (`config` command)

Use the `gossh config` subcommands to manage your connection configurations.

- **Add a new connection**:
    ```sh
    gossh config add [flags]
    ```
    **Modes**:
    - **Interactive Mode**: `gossh config add -i` or `gossh config add --interactive`
    - **Flag-based Mode**:
        ```sh
        gossh config add -n <name> -u <user> -H <host> [other-flags...]
        ```
    **Flags**:
    - `-i, --interactive`: Use interactive mode to add a new connection.
    - `-n, --name`: The connection name (required in non-interactive mode).
    - `-u, --user`: The username (required in non-interactive mode).
    - `-H, --host`: The host address (required in non-interactive mode).
    - `-g, --group`: Assign the connection to a group.
    - `-p, --port`: The port number (defaults to 22).
    - `-k, --key`: Path to the private key.
    - `-P, --use-password`: Use a saved password by its alias for authentication.

- **List saved connections**:
    ```sh
    gossh config list [flags]
    ```
    **Flags**:
    - `-g, --group`: Filter connections by group name.

- **Remove a connection**:
    ```sh
    gossh config remove <connection-name>
    ```

### Connecting & Executing

- **Connect to a server (Interactive Session)**:
    ```sh
    gossh connect <connection-name>
    ```

- **Execute a remote command**:
    ```sh
    gossh exec <connection-name> "<your-command>"
    ```
    Example: `gossh exec web-server "sudo systemctl status nginx"`

- **Test a connection**:
    ```sh
    gossh test <connection-name>
    ```
    This command attempts to authenticate and then immediately disconnects to verify the configuration.

### Managing Passwords (`password` command)

Securely store and manage passwords for your connections.

- **Add a new password**:
    ```sh
    gossh password add <alias>
    ```
    You will be prompted to enter the password.

- **List saved password aliases**:
    ```sh
    gossh password list
    ```

- **Remove a password**:
    ```sh
    gossh password remove <alias>
    ```

### Managing Groups

- **List all groups**:
    ```sh
    gossh groups
    ```

### Importing & Exporting

- **Export configuration**:
    ```sh
    gossh config export > gossh_backup.json
    ```
    This prints all connections and encrypted credentials to standard output.

- **Import configuration**:
    ```sh
    gossh config import <file-path>
    ```
    Imports a configuration from a file created by the `export` command.
    **Warning**: This will overwrite your existing configuration.
    **Flags**:
    - `-f, --force`: Skip the confirmation prompt before overwriting.

## Configuration Files

Configuration files are stored in `~/.config/gossh/`:

- `config.json`: Stores connection configurations.
- `credentials.json`: Stores encrypted passwords.
- `secret.key`: The encryption key for your passwords.

**Note**: Do not share your `secret.key` or `credentials.json` files as they contain sensitive information.

---
This project was created by pi12138 with the assistance of Google's Gemini.