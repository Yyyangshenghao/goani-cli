# 安装指南

## Windows

### 方式一：下载二进制（推荐）

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载 `goani-windows-amd64.zip`

2. 解压得到 `goani.exe`

3. 将 `goani.exe` 移动到一个固定目录，例如 `C:\Tools\goani\`

4. **添加到环境变量 PATH**

   **方法 A：通过系统设置（推荐）**
   - 右键「此电脑」→「属性」→「高级系统设置」
   - 点击「环境变量」
   - 在「系统变量」中找到 `Path`，点击「编辑」
   - 点击「新建」，添加 `goani.exe` 所在目录
   - 点击「确定」保存

   **方法 B：PowerShell**
   ```powershell
   # 添加到用户 PATH（将路径替换为你的实际路径）
   $env:Path += ";C:\Tools\goani"
   [Environment]::SetEnvironmentVariable("Path", $env:Path, [EnvironmentVariableTarget]::User)
   ```

5. **重启终端**，然后验证：
   ```powershell
   goani version
   ```

### 方式二：Go install

需要先安装 Go 1.22+：

```powershell
go install github.com/Yyyangshenghao/goani-cli/cmd/goani@latest
```

---

## macOS

### 方式一：下载二进制（推荐）

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载对应版本：
   - Intel Mac：`goani-darwin-amd64.tar.gz`
   - Apple Silicon (M1/M2/M3)：`goani-darwin-arm64.tar.gz`

2. 解压并安装：
   ```bash
   tar -xzf goani-darwin-arm64.tar.gz
   sudo mv goani /usr/local/bin/
   chmod +x /usr/local/bin/goani
   ```

3. 验证：
   ```bash
   goani version
   ```

### 方式二：Go install

```bash
go install github.com/Yyyangshenghao/goani-cli/cmd/goani@latest
```

### 方式三：Homebrew（计划中）

```bash
brew install yyyangshenghao/tap/goani
```

---

## Linux

### 方式一：下载二进制（推荐）

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载对应版本：
   - x86_64：`goani-linux-amd64.tar.gz`
   - ARM64：`goani-linux-arm64.tar.gz`

2. 解压并安装：
   ```bash
   tar -xzf goani-linux-amd64.tar.gz
   sudo mv goani /usr/local/bin/
   chmod +x /usr/local/bin/goani
   ```

3. 验证：
   ```bash
   goani version
   ```

### 方式二：Go install

```bash
go install github.com/Yyyangshenghao/goani-cli/cmd/goani@latest
```

### 方式三：AUR（Arch Linux）

```bash
yay -S goani-cli-bin
```
