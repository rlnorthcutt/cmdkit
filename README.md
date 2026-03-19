# cmdkit

Lightweight Go library for building simple CLI tools. Provides a colored logger, interactive config resolution (Flag > Env > Prompt > Default), and graceful signal handling. No global state. No external dependencies — stdlib only.

## Packages

### `logger`

Colored, level-aware output. All methods accept a printf-style format string and variadic args.

| Method | Output | Destination |
|--------|--------|-------------|
| `Info(msg, v...)` | cyan `[INFO]` | stdout |
| `Success(msg, v...)` | green `[SUCCESS]` | stdout |
| `Detail(msg, v...)` | plain `-------` | stdout (verbose only) |
| `Debug(msg, v...)` | red `[--dEbUg--]` | stdout (always, dev use only) |
| `Warn(msg, v...)` | yellow `[WARNING]` | stderr |
| `Error(msg, v...)` | red `[ERROR]` | stderr |
| `Fatal(msg, v...)` | magenta `[FATAL]` | stderr, then `os.Exit(1)` |

`Detail` only produces output when the logger was created with `logger.New(true)` (verbose mode). `Debug` always prints — use it during development and remove it before shipping.

### `ui`

Interactive prompts, TTY detection, config resolution, and signal handling. All in one session object.

**Config resolution** via `ResolveString` applies this precedence: Flag > Env > Prompt > Default. The value pointer is updated in place with whichever source wins. Pass the flag's current value and a boolean indicating whether it was explicitly set — works with any flag library (cobra, pflag, stdlib `flag`, etc).

**Signal handling** via `WithInterrupt` registers SIGINT/SIGTERM handlers and stores a cancellable context on `ui.Ctx`. Pass `ui.Ctx` to long-running operations so Ctrl+C cancels them cleanly. Call `StopSignal()` (typically via `defer`) to release the handler when the command exits.

**`ui.Ctx` is always safe to use** — it is `context.Background()` until `WithInterrupt` is called, so you can pass it unconditionally without nil-checking.

**`Confirm`** returns `true` without prompting when not interactive, so automation and CI pipelines proceed unblocked. If you need an explicit opt-in, gate on a flag instead.

### `sys`

`ContextWithInterrupt(parent)` returns a context that cancels on SIGINT or SIGTERM. Use this when you need signal-aware cancellation without a `UI` — for example in background workers or daemons. For interactive CLI tools, prefer `ui.WithInterrupt`.

## Usage

```go
func execute(cmd *cobra.Command, args []string) {
    // 1. Init (no globals)
    log := logger.New(verbose)
    userInterface := ui.New(nonInteractive).
        WithLogger(log).
        WithInterrupt(context.Background())
    defer userInterface.StopSignal()

    // 2. Resolve config: flag > env > prompt > default
    // Look up the flag value with whatever library you use — cmdkit is framework-agnostic
    inputVal, _ := cmd.Flags().GetString("input")
    userInterface.ResolveString(inputVal, cmd.Flags().Changed("input"), "TOOL_INPUT", "Enter URL", &cfg.Input)

    // 3. Summarize & confirm
    fmt.Printf("Target: %s\n", cfg.Input)
    if !userInterface.Confirm("Proceed?") {
        return
    }

    // 4. Pass ui.Ctx to long-running work so Ctrl+C cancels it
    log.Info("Working...")
    doWork(userInterface.Ctx, cfg)
}
```

Tools using cmdkit should support a `--dry-run` (or similar) flag and use the logger to report what would have been done instead of performing destructive actions.
