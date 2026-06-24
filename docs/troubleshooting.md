# Troubleshooting

## hook.log empty

1. Restart IDE after `slim setup`
2. Use **agent** mode (not ask-only) and run a shell tool
3. `slim --target cursor doctor` — hooks module should be OK
4. Check `~/.cursor/hooks.json` contains `slim internal hook`

## Disable one feature

```bash
slim --target cursor config set bash_compress off
```

## Remove hooks only

```bash
slim --target cursor off --module hooks
```

## Copilot vs Cursor

Always pass `--target`:

| Target | Data home |
|--------|-----------|
| copilot | `~/.copilot/slim/` |
| cursor | `~/.cursor/slim/` |

## RTK optional module

```bash
slim setup --module hooks-rtk   # requires rtk in PATH
```
