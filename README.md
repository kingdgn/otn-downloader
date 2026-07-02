# OTN Downloader

OTN Downloader 是一个基于二维码的单向离线文件传输工具。发送端把文件切片编码成二维码流或二维码图片，接收端用手机浏览器扫码、上传图片并在浏览器内合并下载文件。

本项目基于 [alingse/otn-downloader](https://github.com/alingse/otn-downloader) 改造，新增了图片导出、批量图片识别、缺片补播和完整缺片列表显示。

## 在线接收页

手机端优先使用公共 HTTPS 页面：

```text
https://kingdgn.github.io/otn-downloader/
```

摄像头扫码必须运行在 HTTPS 或 localhost 等安全上下文中。使用上面的 GitHub Pages 地址时，手机不需要安装本地证书，也不需要访问电脑的局域网地址。

如果你 fork 了本项目，请在 GitHub Pages 开启后把地址替换为：

```text
https://你的用户名.github.io/仓库名/
```

## 快速开始

### 方式一：下载预编译版本

不想安装 Go 的用户，直接下载 GitHub Releases 里的可执行文件：

```text
https://github.com/kingdgn/otn-downloader/releases/latest
```

Windows 用户可下载：

```text
otn-downloader-windows-amd64.exe
```

然后在 PowerShell 中运行：

```powershell
.\otn-downloader-windows-amd64.exe encode --fps 8 --loop 3 --chunk-size 120 -f example.zip
```

CentOS / Linux amd64 用户可下载：

```text
otn-downloader-centos-amd64
```

然后执行：

```bash
chmod +x ./otn-downloader-centos-amd64
./otn-downloader-centos-amd64 encode --fps 8 --loop 3 --chunk-size 120 -f example.zip
```

### 方式二：用 Go 安装

如果已经安装 Go，也可以安装命令行工具：

```bash
go install github.com/kingdgn/otn-downloader@main
```

### 接收与发送

手机打开在线接收页：

```text
https://kingdgn.github.io/otn-downloader/
```

点击“开始”，允许摄像头权限。

电脑端发送文件：

```bash
otn-downloader encode --fps 8 --loop 3 --chunk-size 120 -f example.zip
```

参数含义：

| 参数 | 含义 |
|------|------|
| `--fps 8` | 终端二维码每秒切换 8 张。过高会增加漏扫概率。 |
| `--loop 3` | 完整二维码序列重复播放 3 轮。 |
| `--chunk-size 120` | 每张数据二维码承载 120 字节原始文件切片。越大则帧数越少，但二维码更复杂。推荐最大参数不超过512。 |
| `-f example.zip` | 要传输的源文件。 |

## 缺片补播

摄像头扫码时，一轮下来通常只会漏掉少量二维码。接收页会显示“完整缺失序列”，可以复制后只补播缺失切片。

```bash
otn-downloader encode --fps 8 --loop 20 --chunk-size 120 -f example.zip --skip-meta -s "0 37 40 46 67"
```

也可以使用语义更明确的参数：

```bash
otn-downloader encode --fps 8 --loop 20 --chunk-size 120 -f example.zip --skip-meta --missing-slices "0,37，40 46 67-70"
```

补片注意：

- `-f` 文件必须和第一轮完全一致。
- `--chunk-size` 必须和第一轮完全一致。
- 使用 `--skip-meta` 时，手机页面不要刷新、不要点“清空”。
- `-s` / `--missing-slices` 支持空格、英文逗号、中文逗号和区间，例如 `67-70`。

## 交互补片模式

如果想一轮一轮手动输入缺片列表，可以使用：

```bash
otn-downloader encode --fps 8 --loop 1 --chunk-size 120 -f example.zip --interactive-missing
```

流程：

1. 第一轮播放完整序列。
2. 手机接收页显示缺片列表。
3. 终端提示输入下一轮缺片序列。
4. 输入后只播放这些缺片。
5. 直接回车退出。

## 导出二维码图片

如果不想用终端动态二维码，也可以直接生成 PNG 图片序列：

```bash
otn-downloader encode --chunk-size 120 -f example.zip --output-dir qr-out-example --image-scale 12
```

输出示例：

```text
qr-out-example/
├── manifest.json
├── manifest.png
├── frame_000000.png
├── frame_000001.png
└── ...
```

然后在接收页选择“二维码图片”，上传 `manifest.png` 和所有 `frame_*.png` 即可合并下载。

## 推荐参数

| 场景 | 推荐命令 |
|------|------|
| 稳定扫码 | `--fps 5 --chunk-size 60` |
| 平衡速度与成功率 | `--fps 8 --chunk-size 120` |
| 图片上传模式 | `--chunk-size 120 --image-scale 12` |
| 大文件探索 | 从 `--chunk-size 180` 或 `240` 开始测试 |

600KB 文件如果使用 `--chunk-size 60`，会产生约一万张二维码；如果改为 `--chunk-size 240`，理论切片数可降到约 2500 张，但单张二维码更复杂，识别成功率需要按设备测试。

## 本地开发

```bash
git clone https://github.com/kingdgn/otn-downloader.git
cd otn-downloader
go test ./...
go build .
```

本地预览接收页：

```bash
python -m http.server 18081
```

电脑浏览器可打开：

```text
http://127.0.0.1:18081/
```

注意：手机访问电脑的局域网 HTTP 地址时，摄像头会被浏览器安全策略禁用。手机扫码请使用 GitHub Pages 的 HTTPS 页面。

## GitHub Pages 部署

仓库发布后，在 GitHub 中打开：

```text
Settings -> Pages -> Build and deployment -> Source: Deploy from a branch
```

选择：

```text
Branch: main
Folder: / (root)
```

保存后，接收页地址为：

```text
https://kingdgn.github.io/otn-downloader/
```

## 协议格式

当前协议保持与原项目兼容：

```text
m:json:{"filename":"example.zip","total":100,"file_size":12000,"chunk_size":120}
d:<index>:<base64>
```

- `m:json` 是文件元信息。
- `d:<index>` 是第 `index` 个文件切片。
- 接收端收齐所有切片后，在浏览器内 Base64 解码并触发下载。

## License

本项目沿用原项目许可证。详见 [LICENSE](LICENSE)。
