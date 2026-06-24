# Onboarding (4 weeks)

## Week 1 — Install

```bash
make build
./bin/slim --target cursor setup --all
./bin/slim --target cursor doctor
```

Copy `templates/repo/` into a project for per-repo savings.

## Week 2 — Habits

- Ask mode for questions; Agent for multi-step
- Don't switch models mid-thread
- Disable unused MCP servers

## Week 3 — Tune

```bash
./bin/slim --target cursor config list
./bin/slim --target cursor report --days 7
```

## Week 4 — Team

Share `docs/practices.md` patterns; opt-in repo templates only.
