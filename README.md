# gossh

`gossh` is a command-line tool written in Go that provides a simple and secure way to manage and connect to your SSH servers. It allows you to save connection configurations, manage encrypted passwords, and easily connect to your servers with a simple command.

## Features

- **Connection Management**: Add, list, and remove SSH connection configurations.
- **Secure Password Storage**: Passwords are encrypted using AES-256 and stored locally.
- **Multiple Authentication Methods**: Supports password, private key, and interactive password authentication.
- **Easy to Use**: Simple and intuitive command-line interface.

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

- **Add a new connection**:
    ```sh
    gossh add -n <connection-name> -u <user> -H <host> [-p <port>] [-k <key-path>] [-P <password-alias>]
    ```
    - `-n`: Connection name (required)
    - `-u`: Username (required)
    - `-H`: Host address (required)
    - `-p`: Port number (defaults to 22)
    - `-k`: Path to your private key
    - `-P`: Use a saved password by its alias

- **List saved connections**:
    ```sh
    gossh list
    ```

- **Remove a connection**:
    ```sh
    gossh remove <connection-name>
    ```

### Managing Passwords

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

### Connecting to a Server

- **Connect to a saved connection**:
    ```sh
    gossh connect <connection-name>
    ```

## Configuration

Configuration files are stored in `~/.config/gossh/`:

- `config.json`: Stores connection configurations.
- `credentials.json`: Stores encrypted passwords.
- `secret.key`: The encryption key for your passwords.

**Note**: Do not share your `secret.key` or `credentials.json` files.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.