# DeerFlow - Deep Research Agent (Go Port)

This is a Go implementation of the [ByteDance DeerFlow](https://github.com/bytedance/deer-flow) deep research agent, built using [langgraphgo](https://github.com/smallnest/langgraphgo) and [langchaingo](https://github.com/tmc/langchaingo).

DeerFlow is a multi-agent system designed to conduct deep research on a given topic. It plans a research strategy, executes search steps (simulated or real), and synthesizes a comprehensive report.

## Features

- **Multi-Agent Architecture**: Orchestrates `Planner`, `Researcher`, and `Reporter` agents using a state graph.
- **Web Interface**: A modern, dark-themed web UI with real-time status updates using Server-Sent Events (SSE).
- **CLI Support**: Can be run directly from the command line for quick queries.
- **Extensible**: Built on `langgraphgo`, making it easy to add new nodes, tools, or complex control flows.

## Prerequisites

- Go 1.23 or higher
- An OpenAI-compatible API Key (e.g., OpenAI, DeepSeek)

## Configuration

Set the following environment variables:

```bash
export OPENAI_API_KEY="your-api-key"

# Optional: If using DeepSeek or another compatible provider
export OPENAI_API_BASE="https://api.deepseek.com/v1" 
```

## Usage

### Web Interface (Recommended)

Build and run the application:

```bash
go build -o deerflow ./showcases/deerflow
./deerflow
```

Open your browser and navigate to `http://localhost:8080`.

### Command Line Interface (CLI)

Run a query directly from the terminal:

```bash
./deerflow "What are the latest advancements in solid state batteries?"
```

## Project Structure

- **`main.go`**: Entry point. Handles CLI arguments and starts the HTTP server.
- **`graph.go`**: Defines the `State` struct and the Graph topology (Nodes and Edges).
- **`nodes.go`**: Contains the implementation logic for `Planner`, `Researcher`, and `Reporter`.
- **`web/`**: Contains the frontend assets (HTML, CSS, JS).

## Architecture

The agent follows a sequential workflow:

1.  **Planner**: Decomposes the user's query into a step-by-step research plan.
2.  **Researcher**: Iterates through the plan, gathering information for each step.
3.  **Reporter**: Synthesizes all gathered information into a final, formatted report.

## License

MIT
