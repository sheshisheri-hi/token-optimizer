# token-optimizer

GitHub Copilot token optimizer refers to tools, community scripts, and workflow strategies that reduce GitHub AI Credit consumption after the move to usage-based billing.

## Why optimization matters

Each Copilot prompt can spend a large portion of its token window on hidden context such as:

- open editor tabs
- related repository snippets
- command output
- repository metadata

That means optimization is not only about writing shorter prompts; it is also about reducing passive context.

## Active tools and open source projects

### `alexgreensh/token-optimizer`

A local, single-file tool that focuses on:

- compressing command output
- trimming bloated configuration payloads
- removing duplicate system prompts
- exposing a local HTML dashboard for token and cost tracking

### `olivomarco/github-copilot-token-optimization`

A community guide and repository that documents practical workarounds, including the Runtime Token Killer hook, which filters verbose terminal output before it enters Copilot agent context.

### `githubnext/agentic-ops`

An experimental workflow ecosystem with a `copilot-token-optimizer.md` blueprint for auditing workflow logs over a rolling 7-day window, identifying expensive tasks, and surfacing optimization opportunities.

## Strategic repository practices

To reduce token usage directly inside a repository:

1. Use lower-cost models for routine scaffolding and boilerplate tasks.
2. Add heavy artifacts, minified assets, mock data, and lockfiles to `.copilotignore` when they do not need to be indexed.
3. Close unrelated editor tabs before asking broad architectural questions.
4. Keep Copilot instructions, agent definitions, and repository rules concise.

## Choosing an optimization approach

- For CI or workflow spend, focus on workflow-log auditing and automation.
- For developer AI Credit usage in editors, focus on context reduction, model choice, and prompt compression.