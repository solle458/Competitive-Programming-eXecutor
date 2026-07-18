# Competitive-Programming-eXecutor

AtCoder 向けの競技プログラミング CLI ツール。

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/solle458/Competitive-Programming-eXecutor/main/install.sh | sh
```

## Quick Start

```bash
cpx init
cpx setup abc464
cd abc464
cpx test a
cpx merge a
cpx submit a   # 開催中コンテストのみ
```

## Configuration

`cpx init` でワークスペースルートに `.cpx/config.yaml` が作成される。  
`init` 以外のコマンドは、カレントディレクトリから親ディレクトリへ `.cpx/config.yaml` を探索して読み込む。

```yaml
file:
  root_dir: /path/to/workspace
  library_dirs:
    - /path/to/workspace/library
  default_lang: cpp
  atcoder_session: ""   # 任意。開催中コンテストの setup / submit で使用
```

| 項目 | 説明 |
|------|------|
| `root_dir` | ワークスペースルート |
| `library_dirs` | `merge` で検索するライブラリディレクトリ |
| `default_lang` | `merge` の `--lang` 未指定時のデフォルト |
| `atcoder_session` | AtCoder の `REVEL_SESSION` cookie 値 |

`atcoder_session` は `ATCODER_SESSION` 環境変数でも指定可能（環境変数が優先）。

## Directory Layout

`cpx setup <contest-id>` 後のディレクトリ構造:

```
<workspace>/
  .cpx/config.yaml
  library/              # init で作成
  <contest-id>/
    a/
      main.cpp
      submission.cpp    # merge / submit で生成
      a.out             # test で生成
      test/
        sample-1.in
        sample-1.out
        sample-1.test   # test で生成
    b/
    ...
```

- 問題ディレクトリ名は問題インデックスの小文字（`A` → `a`）
- `test` / `merge` / `submit` の `<problem-path>` は `a` または `<contest-id>/a` 形式

## Commands

### `cpx init`

ワークスペースを初期化する。config 読み込みは不要。

| | |
|---|---|
| 引数 | なし |
| フラグ | なし |

**動作:**

- `.cpx/config.yaml` を作成（`library_dirs: [<root>/library]`, `default_lang: cpp`）
- `.cpx/templates/source/source_template.cpp` を作成
- `.cpx/templates/library/library_template.hpp` を作成
- `library/` ディレクトリは作成しない（config にパスのみ記載）

---

### `cpx setup <contest-id>`

コンテストの問題ディレクトリとサンプルケースを準備する。

| | |
|---|---|
| 引数 | `<contest-id>`（例: `abc464`） |
| フラグ | `-l, --lang` 言語（default: `cpp`） |

**動作:**

- AtCoder コンテストの存在確認（`Supports`）
- 問題一覧取得（kenkoooo API → 失敗時は AtCoder tasks ページ）
- 各問題に対して並列で:
  - `{cwd}/{contest-id}/{index}/` を作成
  - `main.{lang}` がなければテンプレートを配置
  - `test/` がなければサンプル入出力をスクレイピングして `sample-N.in/out` を保存

**注意:**

- 開催中コンテストのサンプル取得には `atcoder_session` が必要な場合がある
- サンプルが見つからない問題はスキップして続行

---

### `cpx test <problem-path>`

サンプルケースでローカルテストを実行する。

| | |
|---|---|
| 引数 | `<problem-path>`（例: `a`, `abc464/a`） |
| フラグ | `-l, --lang` 言語（default: `cpp`） |
| | `-t, --time-limit` 制限時間秒（default: `2`） |

**動作:**

1. コンパイル（C++: `g++ -std=c++20 -O3`, Python: 存在確認のみ）
2. 各 `test/*.in` を stdin に渡して実行し `test/*.test` に出力を保存
3. `test/*.test` と `test/*.out` を比較（AC / WA / TLE を表示）

**終了コード:** 全ケース AC なら 0、WA / TLE なら 1

---

### `cpx merge <problem-path>`

ライブラリを展開して提出用ファイルを生成する。

| | |
|---|---|
| 引数 | `<problem-path>` |
| フラグ | `-l, --lang` 言語（default: 空 → `config.default_lang`） |

**動作:**

- `{problem-path}/main.{lang}` を読み込み
- `/* -- libraries --*/` マーカー以降の `#include "..."` から `.hpp` ライブラリを BFS 展開
- `{problem-path}/submission.{lang}` を生成

**要件:**

- `main.{lang}` に `/* -- libraries --*/` マーカーが必要
- 各 `.hpp` に `/* -- library code --*/` マーカーが必要

---

### `cpx submit <problem-path>`

テスト・マージ・提出を一括実行する。**開催中コンテスト専用。**

| | |
|---|---|
| 引数 | `<problem-path>` |
| フラグ | `-l, --lang` 言語（default: `cpp`） |
| | `-t, --time-limit` テスト制限時間秒（default: `2`） |
| | `--skip-test` テストをスキップして merge + submit のみ |
| | `-c, --copy` 提出せずマージ結果をクリップボードへコピー（過去問向け） |

**動作:**

1. 問題ディレクトリの存在確認
2. `cpx test` 相当（`--skip-test` でスキップ可）
3. `cpx merge` 相当で `submission.{lang}` 生成
4. `--copy` の場合はクリップボードへコピーして終了
5. それ以外は `atcoder_session` を [online-judge-tools](https://github.com/online-judge-tools/oj) の cookie.jar に同期し、**MiB 対応パッチ付き**で `oj submit` 相当を実行

素の `oj submit` は AtCoder のメモリ制限表記（`MiB` / `KiB`）をパースできず落ちることがある。`cpx submit` は oj 実行前にパッチを当てるため、その問題を回避する。oj のログに出る `#include_<iostream>` のような `_` は表示用の置換で、提出ファイル自体は壊れていない。

**要件:**

- `oj` がインストール済み（`pip install online-judge-tools` または `uv tool install online-judge-tools`）
- `atcoder_session` または `ATCODER_SESSION` が設定済み（`--copy` 時は不要）
- **開催中コンテストのみ** — 過去問は `cpx submit <problem> --copy` のあとブラウザから提出

**task URL 解決:**

| 引数 | contest ID | index |
|------|-----------|-------|
| `abc464/a` | `abc464` | `a` |
| `a`（cwd がコンテスト dir） | `basename(cwd)` | `a` |

---

## Typical Workflow

```
cpx init
cpx setup abc464 -l cpp
cd abc464

# 問題 A
vim a/main.cpp
cpx test a
cpx merge a
cpx submit a        # 本番提出（開催中のみ）

# ライブラリ追加後
cpx merge a         # submission.cpp 再生成
```

## Dependencies

| コマンド | 外部依存 |
|---------|---------|
| `test` (cpp) | `g++` |
| `test` (py) | `python3` |
| `setup` | ネットワーク（AtCoder / kenkoooo） |
| `submit` | `oj`, `g++` or `python3`, `REVEL_SESSION` |
