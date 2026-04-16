# iconkit

<p align="center">
  <strong>一个面向开发者的图标处理 CLI 工具。</strong><br />
  用一条命令完成多尺寸导出、圆角、留白、背景填充，以及 <code>favicon.ico</code> 生成。
</p>

<p align="center">
  <a href="./README.md">English</a>
</p>

<p align="center">
  <a href="https://github.com/Tendo33/iconkit/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/Tendo33/iconkit/ci.yml?branch=main&label=CI&logo=githubactions" alt="CI" />
  </a>
  <a href="https://github.com/Tendo33/iconkit/releases">
    <img src="https://img.shields.io/github/v/release/Tendo33/iconkit?display_name=tag&logo=github" alt="Release" />
  </a>
  <img src="https://img.shields.io/badge/Go-1.26.2-00ADD8?logo=go" alt="Go Version" />
  <a href="./LICENSE">
    <img src="https://img.shields.io/github/license/Tendo33/iconkit" alt="License" />
  </a>
</p>

## 亮点

- 从一张 PNG 或 JPG 快速生成多尺寸图标
- 支持圆角，并按输出尺寸自动缩放圆角半径
- 支持为图标增加留白，适合安全区和 maskable icon
- 支持填充透明区域背景色
- 可同时导出多尺寸 `favicon.ico`
- 支持整目录批量处理

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [用法](#用法)
- [参数说明](#参数说明)
- [预设尺寸](#预设尺寸)
- [favicon.ico](#faviconico)
- [留白与背景](#留白与背景)
- [JSON 配置](#json-配置)
- [批量处理](#批量处理)
- [输出结果](#输出结果)
- [开发](#开发)
- [发布](#发布)
- [许可证](#许可证)

## 安装

### 一行命令安装（macOS / Linux）

```bash
curl -fsSL https://raw.githubusercontent.com/Tendo33/iconkit/main/install.sh | sh
```

### Homebrew（仅 macOS）

```bash
brew install --cask Tendo33/homebrew-tap/iconkit
```

当前 `iconkit` 分发的是未签名的 macOS 二进制文件。如果安装后被 macOS 阻止运行，可执行：

```bash
xattr -dr com.apple.quarantine "$(brew --prefix)/Caskroom/iconkit/latest/iconkit"
```

### Go 安装

```bash
go install github.com/Tendo33/iconkit@latest
```

Linux 用户请使用一键安装脚本、`go install`，或从 Releases 下载二进制文件。

### 二进制下载

可从 [GitHub Releases](https://github.com/Tendo33/iconkit/releases) 下载最新构建产物。

## 快速开始

```bash
# 默认输出 16、32、64、128
iconkit icon.png

# 使用 web 预设并生成 favicon.ico
iconkit icon.png -p web --ico

# 加圆角并输出到自定义目录
iconkit icon.png -r 20 -o ./dist

# 增加留白并填充白色背景
iconkit icon.png --pad 0.1 --bg "#ffffff"
```

## 用法

```bash
iconkit [input] [options]
```

`<input>` 可以是单个 `.png`、`.jpg`、`.jpeg` 文件，也可以是包含图片的目录。

### 示例

```bash
# 默认生成 16、32、64、128 像素图标
iconkit icon.png

# 指定尺寸并加圆角
iconkit icon.png -r 20 -s 16,32,64,128

# Web 预设（favicon 常用尺寸）
iconkit icon.png -p web

# Chrome 扩展图标
iconkit icon.png -p chrome-ext

# Firefox 扩展图标
iconkit icon.png -p firefox-ext

# iOS AppIcon 尺寸
iconkit icon.png -p ios

# Android mipmap 图标
iconkit icon.png -p android

# PWA 图标
iconkit icon.png -p pwa

# 导出 PNG 的同时生成 favicon.ico
iconkit icon.png -p web --ico

# 增加 10% 留白并填充白色背景
iconkit icon.png --pad 0.1 --bg "#ffffff"

# 用指定颜色填充透明区域
iconkit icon.png --bg "#1a1a2e" -p chrome-ext

# 批量处理目录里的所有图片
iconkit ./assets/ -p web

# 自定义输出目录并强制覆盖
iconkit icon.png -s 16,32 -o ./dist -f

# 使用 JSON 配置文件
iconkit icon.png -c iconkit.json
```

## 参数说明

| 参数 | 短参数 | 说明 | 默认值 |
|------|--------|------|--------|
| `--sizes` | `-s` | 输出尺寸，逗号分隔 | `16,32,64,128` |
| `--radius` | `-r` | 圆角半径，单位像素 | `0` |
| `--preset` | `-p` | 使用下方预设尺寸 | 无 |
| `--out` | `-o` | 输出目录 | `./icons` |
| `--force` | `-f` | 覆盖已存在文件 | `false` |
| `--config` | `-c` | JSON 配置文件路径 | 自动检测 `iconkit.json` |
| `--pad` |  | 留白比例，范围 `0.0` 到 `0.5` | `0` |
| `--bg` |  | 十六进制背景色，如 `#ffffff` | 透明 |
| `--ico` |  | 同时生成尺寸 `<= 256` 的 `favicon.ico` | `false` |
| `--version` | `-v` | 输出版本号 | 无 |

当指定 `-p` 时，`-s` 会被忽略。

## 预设尺寸

| 名称 | 尺寸 | 使用场景 |
|------|------|----------|
| `web` | 16, 32, 48, 64, 128, 256 | favicon 与 PWA 图标 |
| `chrome-ext` | 16, 32, 48, 128 | Chrome Extension（Manifest V3） |
| `firefox-ext` | 32, 48, 64, 96, 128 | Firefox Add-on |
| `pwa` | 192, 512 | Progressive Web App |
| `ios` | 20, 29, 40, 58, 60, 76, 80, 87, 120, 152, 167, 180, 1024 | iOS AppIcon |
| `android` | 36, 48, 72, 96, 144, 192, 512 | Android mipmap 与 Play Store |

## favicon.ico

使用 `--ico` 可以在导出 PNG 的同时生成一个多尺寸 `.ico` 文件：

```bash
iconkit icon.png -p web --ico
```

最终写入 `.ico` 的尺寸仅包含 `<= 256` 的输出结果。

## 留白与背景

使用 `--pad` 为原始图标增加留白：

```bash
iconkit icon.png --pad 0.1 -p ios
```

使用 `--bg` 为透明区域填充纯色背景：

```bash
iconkit icon.png --bg "#ffffff" -p android
```

两者可以组合使用：

```bash
iconkit icon.png --pad 0.1 --bg "#1a1a2e" -r 20 -p web --ico
```

## JSON 配置

可在项目根目录创建 `iconkit.json`：

```json
{
  "input": "icon.png",
  "sizes": [16, 32, 64, 128],
  "radius": 20,
  "preset": "web",
  "out": "./dist",
  "force": true,
  "pad": 0.1,
  "bg": "#112233",
  "ico": true
}
```

当前 JSON 配置支持 `input`、`sizes`、`radius`、`preset`、`out`、`force`、`pad`、`bg`、`ico`。

如果在 `iconkit.json` 里设置了 `input`，执行 `iconkit` 时可以不再额外传位置参数。

命令行参数的优先级始终高于配置文件。

## 批量处理

传入目录时，会处理其中所有 `.png`、`.jpg`、`.jpeg` 文件：

```bash
iconkit ./assets/ -s 32,64
```

输出文件会保留原始文件名作为前缀：

```text
logo-32.png
logo-64.png
badge-32.png
badge-64.png
```

批量模式下如果启用 `--ico`，每张源图还会额外生成一个以原文件名命名的 `.ico` 文件。

## 输出结果

单文件输入时，输出结构类似：

```text
./icons/
|- icon-16.png
|- icon-32.png
|- icon-48.png
|- icon-64.png
|- icon-128.png
|- icon-256.png
`- favicon.ico
```

批量输入时，文件名格式为 `{name}-{size}.png` 和 `{name}.ico`。

## 开发

```bash
go test ./... -v
go build -o iconkit .
```

## 发布

```bash
git tag v2.1.0
git push origin v2.1.0
```

发布流程由 GoReleaser 与 GitHub Actions 自动完成。

## 许可证

MIT
