# Annict Slack Bot

Slack で `@<bot名> annict today` とメンションすると、Annict API からその日に放送予定のアニメ情報を取得して、整形されたメッセージで返信する Go の Slack Bot です。

## 概要

指定されたコマンドに反応し、Annict の GraphQL API を利用して情報を取得し、Slack の Block Kit を使って見やすい形式でチャンネルに投稿します。

## 必要なもの

- **Go 環境:** Go 1.18 以上を推奨します。
- **Git:** リポジトリのクローンや依存関係の取得に必要です。
- **Slack Bot Token:**
  - Slack アプリを作成し、`Bot Token` (`xoxb-...`形式) を取得します。
  - 必要な権限スコープ: `app_mentions:read`, `chat:write`
- **Slack App-Level Token:**
  - Slack アプリ設定の "Socket Mode" を有効にし、`App-Level Token` (`xapp-...`形式) を生成します。
  - 必要なスコープ: `connections:write`
- **Annict Personal Access Token:**
  - Annict の開発者設定ページ (<https://annict.jp/settings/apps>) から `個人用アクセストークン` を生成します。

## セットアップ

1.  **リポジトリのクローン:**
    ```bash
    git clone https://github.com/monchh/annict-slack-bot.git
    cd annict-slack-bot
    ```
2.  **依存関係の取得:**
    ```bash
    # Makefileを使用する場合
    make deps
    # または直接goコマンドを使用する場合
    # go mod download
    ```
3.  **.env ファイルの作成:**
    プロジェクトのルートディレクトリに `.env` という名前のファイルを作成し、取得したトークンを記述します。

    ```.env
    # Slack API Tokens
    SLACK_BOT_TOKEN="xoxb-YOUR_SLACK_BOT_TOKEN"
    SLACK_APP_TOKEN="xapp-YOUR_SLACK_APP_TOKEN"

    # Annict Personal Access Token
    ANNICT_ACCESS_TOKEN="YOUR_ANNICT_PERSONAL_ACCESS_TOKEN"

    # Optional: Set log level (e.g., debug, info, warn, error)
    # LOG_LEVEL="info"
    # Optional: Flag for development-specific behavior
    # IS_DEVELOPMENT="true"
    ```

    **注意:** この `.env` ファイルは機密情報を含むため、Git リポジトリにコミットしないでください (`.gitignore` に含まれています)。

## ビルドと実行

**Makefile を使用する場合:**

1.  **ビルド:**
    ```bash
    make build
    ```
2.  **実行:**
    ```bash
    make run
    ```
    (実行前に `.env` ファイルが存在するか、環境変数が設定されている必要があります)

**go コマンドを直接使用する場合:**

1.  **ビルド:**
    ```bash
    go build -o annict-slack-bot ./cmd/annict-slack-bot
    ```
2.  **実行:**

    ```bash
    ./annict-slack-bot

    # または環境変数を直接設定する場合
    # export SLACK_BOT_TOKEN="xoxb-..."
    # export SLACK_APP_TOKEN="xapp-..."
    # export ANNICT_ACCESS_TOKEN="..."
    # ./annict-slack-bot
    ```

## Slack での使い方

1.  作成した Slack Bot を、情報を投稿したいチャンネルに招待します。
    ```
    /invite @your-bot-name
    ```
    (`@your-bot-name` は作成した Bot の名前に置き換えてください)
2.  そのチャンネルで Bot にメンションしてコマンドを送信します。
    ```
    @your-bot-name annict_today
    ```
3.  Bot が Annict API から情報を取得し、以下のような形式でその日の放送予定と未視聴の情報を投稿します。

    - **ヘッダー:**
      - `:calendar: YYYY-MM-DD 放送予定のアニメ`
      - `:eyes: 未視聴のアニメ`
    - **各アニメ情報:**
      - アニメタイトル (公式サイトへのリンク付き、存在する場合)
      - エピソード番号 (例: `第1話`)
      - エピソードタイトル (存在する場合)
      - 放送チャンネル名
      - 放送時間 (HH:MM)
      - アニメ画像 (画像 URL が有効な場合)

## 設定

以下の環境変数を `.env` ファイルまたは実行環境で設定することで、Bot の挙動を調整できます。

- `LOG_LEVEL`: ログレベル (`debug`, `info`, `warn`, `error` など。デフォルト: `info`)
- `IS_DEVELOPMENT`: 開発モードフラグ (`true` にするとデバッグログが増えることがあります。デフォルト: `false`)

## Annict API クライアントの更新手順

Annict API の仕様が変更された場合などには、GraphQL クライアントコードを最新の状態に更新する必要がある場合があります。
プロジェクトのルートディレクトリで以下のコマンドを実行することで、Annict API の最新の GraphQL スキーマを取得し、クライアントコードを再生成します。

```bash
make make-client
```

## プログラム構成 (アーキテクチャ)

このアプリケーションは、関心事の分離とテスト容易性の向上を目的として、**Onion Architecture** に基づいて設計されています。依存関係は常に内側（ドメイン）から外側（インフラストラクチャ）へと向かい、逆方向の依存は持ちません。

この構成により、例えば Annict API の仕様変更があった場合でも、影響範囲は主に infrastructure/annict と interfaces/repository に限定され、ドメインロジックへの影響を最小限に抑えることができます。同様に、Slack のライブラリを変更する場合も infrastructure/slack や interfaces/presenter の修正が中心となります。

- **依存性の方向:** `cmd` -> `infrastructure` -> `interfaces` -> `domain`。逆方向の依存はありません。
- **インターフェース:** ドメイン層で定義されたインターフェース (`ProgramRepository`, `ImageValidationService`) をインターフェース層で実装します。また、インターフェース層やインフラストラクチャ層でも、下位の具体的な実装に依存しないようにインターフェースを定義しています（例: `AnnictClient`, `HTTPClient`, `HTTPClient`, `UseCaseExecutor`, `ProgramPresenter`）。
- **DI (Dependency Injection):** `main.go` で各層のコンポーネントをインスタンス化し、コンストラクタを通じて依存性を注入しています。これにより、各コンポーネントは具体的な実装ではなく、インターフェースに依存します。
- **関心の分離:**
  - `domain`: ビジネスルールとエンティティ。
  - `usecase`: アプリケーション固有のロジックの調整。
  - `interfaces`: ドメインとインフラの間の変換・適合。
  - `infrastructure`: 外部サービスとの通信、設定、フレームワーク固有の処理。
  - `cmd`: アプリケーションの起動と DI の設定。

### Domain Layer (domain/)

アプリケーションの中心となるビジネスロジックとデータ構造（エンティティ）を定義します。このレイヤーは他のどのレイヤーにも依存しません。

- entity: アプリケーションの核となるデータ構造（Program, Work, Episode, Channel など）を定義します。
- usecase: アプリケーション固有の処理フロー（ユースケース）を定義します。今回の場合、FetchTodaysProgramsUseCase が今日の放送予定を取得するロジックを担当します。データ永続化（リポジトリ）や外部検証（バリデーション）のためのインターフェースもここで定義されます。

### Interfaces Layer (interfaces/)

Domain 層で定義されたインターフェースを実装し、Domain 層と Infrastructure 層の間のアダプター（変換役）として機能します。

- repository: domain/usecase で定義された ProgramRepository インターフェースを実装します。Infrastructure 層の Annict クライアントを利用してデータを取得し、Domain 層のエンティティにマッピングします。
- validator: domain/usecase で定義された ImageValidationService インターフェースを実装します。Infrastructure 層の HTTP クライアントを利用して画像の URL を検証します。
- presenter: Domain 層のエンティティを受け取り、Slack に表示するための形式（Block Kit）に変換します。

### Infrastructure Layer (infrastructure/)

外部の技術的な詳細（データベース、外部 API、フレームワークなど）を担当します。

- annict: Annict GraphQL API との具体的な通信処理を実装します。
- httpclient: 画像 URL 検証のための HTTP クライアント（リダイレクトを追わない設定）を提供します。
- slack: Slack API (Socket Mode) との接続、イベントの受信、メッセージの送信、ユースケースの呼び出しなど、Slack との連携全体を管理します。
- config: 環境変数や .env ファイルからの設定値の読み込みを担当します。

### cmd Layer (cmd/)

アプリケーションのエントリーポイントです。各レイヤーのコンポーネントを初期化し、依存関係を注入（Dependency Injection）してアプリケーション全体を起動します。

### pkg Layer (pkg/)

プロジェクト固有ではない、汎用的なユーティリティ関数などを配置します。

- jst: 日本標準時 (JST) に関する時間処理のヘルパー関数を提供します。
