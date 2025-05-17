# Annict Slack Bot

This is a Go Slack Bot that, when mentioned with `@<bot_name> annict today` in Slack, fetches anime information scheduled to air on that day from the Annict API and replies with a formatted message.

## Overview

The bot reacts to specified commands, utilizes Annict's GraphQL API to retrieve information, and posts it to the channel in an easy-to-read format using Slack's Block Kit.

## Prerequisites

- **Go Environment:** Go 1.18 or higher is recommended.
- **Git:** Required for cloning the repository and fetching dependencies.
- **Slack Bot Token:**
  - Create a Slack app and obtain a `Bot Token` (in `xoxb-...` format).
  - Required permission scopes: `app_mentions:read`, `chat:write`
- **Slack App-Level Token:**
  - Enable "Socket Mode" in your Slack app settings and generate an `App-Level Token` (in `xapp-...` format).
  - Required scope: `connections:write`
- **Annict Personal Access Token:**
  - Generate a `Personal Access Token` from Annict's developer settings page (<https://annict.jp/settings/apps>).

## Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/monchh/annict-slack-bot.git
    cd annict-slack-bot
    ```
2.  **Fetch dependencies:**
    ```bash
    # If using Makefile
    make deps
    # Or using go command directly
    # go mod download
    ```
3.  **Create `.env` file:**
    Create a file named `.env` in the project's root directory and write the obtained tokens.

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

    **Note:** This `.env` file contains sensitive information, so do not commit it to your Git repository (it's included in `.gitignore`).

## Build and Run

**Using Makefile:**

1.  **Build:**
    ```bash
    make build
    ```
2.  **Run:**
    ```bash
    make run
    ```
    (The `.env` file must exist, or environment variables must be set before running)

**Using go command directly:**

1.  **Build:**
    ```bash
    go build -o annict-slack-bot ./cmd
    ```
2.  **Run:**

    ```bash
    ./annict-slack-bot

    # Or if setting environment variables directly
    # export SLACK_BOT_TOKEN="xoxb-..."
    # export SLACK_APP_TOKEN="xapp-..."
    # export ANNICT_ACCESS_TOKEN="..."
    # ./annict-slack-bot
    ```

## How to Use in Slack

1.  Invite the created Slack Bot to the channel where you want to post information.
    ```
    /invite @your-bot-name
    ```
    (Replace `@your-bot-name` with the name of your created Bot)
2.  Mention the Bot in that channel and send the command.
    ```
    @your-bot-name annict_today
    ```
3.  The Bot will fetch information from the Annict API and post the day's broadcast schedule and unwatched anime in the following format:

    - **Headers:**
      - `:calendar: Anime scheduled to air on YYYY-MM-DD`
      - `:eyes: Unwatched Anime`
    - **Each Anime's Information:**
      - Anime title (with a link to the official site, if it exists)
      - Episode number (e.g., `Episode 1`)
      - Episode title (if it exists)
      - Broadcasting channel name
      - Broadcast time (HH:MM)
      - Anime image (if the image URL is valid)

## Configuration

You can adjust the Bot's behavior by setting the following environment variables in the `.env` file or in your execution environment.

- `LOG_LEVEL`: Log level (`debug`, `info`, `warn`, `error`, etc. Default: `info`)
- `IS_DEVELOPMENT`: Development mode flag (setting to `true` may increase debug logs. Default: `false`)

## How to Update the Annict API Client

You may need to update the GraphQL client code when the Annict API specification changes.
By running the following command in the project's root directory, you can fetch the latest GraphQL schema from the Annict API and regenerate the client code.

```bash
make make-client
```

## Program Structure (Architecture)

This application is designed based on **Onion Architecture** to promote separation of concerns and improve testability. Dependencies always flow inwards (from infrastructure to domain), and there are no reverse dependencies.

This structure ensures that, for example, if the Annict API specification changes, the impact is primarily limited to `infrastructure/annict` and `interfaces/repository`, minimizing the effect on the domain logic. Similarly, if the Slack library is changed, modifications will mainly be in `infrastructure/slack` and `interfaces/presenter`.

- **Direction of Dependency:** `cmd` -> `infrastructure` -> `interfaces` -> `domain`. There are no reverse dependencies.
- **Interfaces:** Interfaces defined in the domain layer (`ProgramRepository`, `ImageValidationService`) are implemented in the interfaces layer. Also, interfaces are defined in the interfaces and infrastructure layers to avoid dependency on concrete lower-level implementations (e.g., `AnnictClient`, `HTTPClient`, `UseCaseExecutor`, `ProgramPresenter`).
- **DI (Dependency Injection):** In `main.go`, components of each layer are instantiated, and dependencies are injected through constructors. This allows each component to depend on interfaces rather than concrete implementations.
- **Separation of Concerns:**
  - `domain`: Business rules and entities.
  - `usecase`: Orchestration of application-specific logic.
  - `interfaces`: Transformation/adaptation between domain and infrastructure.
  - `infrastructure`: Communication with external services, configuration, framework-specific processing.
  - `cmd`: Application startup and DI setup.

### Domain Layer (domain/)

Defines the core business logic and data structures (entities) of the application. This layer does not depend on any other layer.

- `entity`: Defines the core data structures of the application (Program, Work, Episode, Channel, etc.).
- `usecase`: Defines application-specific processing flows (use cases). In this case, `FetchTodaysProgramsUseCase` is responsible for the logic of fetching today's broadcast schedule. Interfaces for data persistence (repository) and external validation (validation) are also defined here.

### Interfaces Layer (interfaces/)

Implements interfaces defined in the Domain layer and acts as an adapter (transformer) between the Domain layer and the Infrastructure layer.

- `repository`: Implements the `ProgramRepository` interface defined in `domain/usecase`. It uses the Annict client from the Infrastructure layer to retrieve data and maps it to Domain layer entities.
- `validator`: Implements the `ImageValidationService` interface defined in `domain/usecase`. It uses the HTTP client from the Infrastructure layer to validate image URLs.
- `presenter`: Receives Domain layer entities and transforms them into a format for display on Slack (Block Kit).

### Infrastructure Layer (infrastructure/)

Handles external technical details (databases, external APIs, frameworks, etc.).

- `annict`: Implements the specific communication processing with the Annict GraphQL API.
- `httpclient`: Provides an HTTP client for image URL validation (configured not to follow redirects).
- `slack`: Manages the overall integration with Slack, including connection with the Slack API (Socket Mode), receiving events, sending messages, and invoking use cases.
- `config`: Responsible for loading configuration values from environment variables and the `.env` file.

### cmd Layer (cmd/)

The entry point of the application. It initializes components of each layer, injects dependencies (Dependency Injection), and starts the entire application.

### pkg Layer (pkg/)

Contains generic utility functions that are not specific to the project.

- `jst`: Provides helper functions for time processing related to Japan Standard Time (JST).
