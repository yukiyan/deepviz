# nanogen

Nanobanana 画像生成 CLI ツール。Zig で書かれた高速・軽量なシングルバイナリ。

## Features

- Gemini API (nanobanana) を使った画像生成
- 外部ライブラリゼロ、純粋な Zig 実装
- ~750KB の静的リンクバイナリ
- XDG Base Directory 準拠

## Requirements

- [Zig](https://ziglang.org/) 0.14.0+
- Gemini API キー

## Quick Start

```bash
# ビルド
zig build -Doptimize=ReleaseSmall

# 画像生成
export NANOGEN_API_KEY="your-api-key"
./zig-out/bin/nanogen -p "夕焼けの富士山"
```

## Installation

```bash
zig build -Doptimize=ReleaseSmall
cp zig-out/bin/nanogen ~/.local/bin/
```

## Usage

```
Usage: nanogen [OPTIONS]

Options:
  -p, --prompt <TEXT>         Generation prompt
  -f, --file <PATH>           Read prompt from file
      --model <NAME>          Model name (default: gemini-2.0-flash-preview-image-generation)
      --aspect-ratio <RATIO>  Aspect ratio: 16:9, 4:3, 1:1, 9:16 (default: 16:9)
      --image-size <SIZE>     Image size: 2K, 4K (default: 2K)
  -o, --output <DIR>          Output directory
      --no-open               Don't auto-open generated image
  -v, --verbose               Enable debug logging
      --version               Show version
  -h, --help                  Show this help
```

### Examples

```bash
# テキストプロンプトから生成
nanogen -p "a cat sitting on a cloud"

# ファイルからプロンプトを読み込み
nanogen -f prompt.txt

# アスペクト比と解像度を指定
nanogen -p "infographic about AI" --aspect-ratio 4:3 --image-size 4K

# 出力先を指定、自動オープンを無効化
nanogen -p "sunset" -o /tmp/images --no-open

# デバッグログを有効化
nanogen -p "hello" -v
```

## Configuration

設定は以下の優先度で適用されます（上が優先）:

1. CLI フラグ
2. 環境変数
3. 設定ファイル
4. デフォルト値

### Environment Variables

| 変数名 | 説明 |
|--------|------|
| `NANOGEN_API_KEY` | API キー（必須） |
| `NANOGEN_MODEL` | モデル名 |
| `NANOGEN_ASPECT_RATIO` | アスペクト比 |
| `NANOGEN_IMAGE_SIZE` | 画像サイズ |
| `NANOGEN_OUTPUT_DIR` | 出力ディレクトリ |
| `NANOGEN_AUTO_OPEN` | 自動オープン（`false` / `0` で無効） |

### Config File

`$XDG_CONFIG_HOME/nanogen/config.json`（デフォルト: `~/.config/nanogen/config.json`）

```json
{
  "api_key": "your-api-key",
  "model": "gemini-2.0-flash-preview-image-generation",
  "aspect_ratio": "16:9",
  "image_size": "2K",
  "auto_open": true
}
```

## Output

生成されたファイルは `$XDG_DATA_HOME/nanogen/`（デフォルト: `~/.local/share/nanogen/`）に保存されます。

```
~/.local/share/nanogen/
├── images/          # 生成された PNG 画像
├── responses/       # API レスポンス JSON
└── logs/            # 実行ログ
```

## Development

```bash
# デバッグビルド
zig build

# テスト実行
zig build test

# リリースビルド（最小サイズ）
zig build -Doptimize=ReleaseSmall

# リリースビルド（最高速度）
zig build -Doptimize=ReleaseFast
```

### Project Structure

```
src/
├── main.zig           # エントリポイント
├── cli.zig            # 引数パーサー
├── config.zig         # 設定管理
├── client.zig         # HTTP/TLS クライアント
├── api.zig            # Gemini API
├── json_build.zig     # JSON リクエスト生成
├── json_parse.zig     # JSON レスポンス解析
├── base64.zig         # Base64 デコード
├── fs.zig             # ファイル I/O・XDG パス
├── opener.zig         # ファイルオープン
├── log.zig            # ロガー
└── timestamp.zig      # タイムスタンプ生成
```

## License

MIT
