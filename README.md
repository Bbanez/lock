# Lock

Simple CLI utility for encrypting and decrypting files using AES-256 encryption in a directory.

## Usage

- To encrypt current directory: `lock`
- To encrypt specific directory relative to current directory: `lock -i path/to/sub_dir`
- To decrypt current directory: `lock -u true`
- To decrypt specific directory relative to current directory: `lock -u true -i path/to/sub_dir`
- Absolute paths are not supported, only relative paths from current directory.
