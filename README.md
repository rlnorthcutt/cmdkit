# cmdkit

Lightweight Go library for building simple CLI tools. Provides a colored logger, interactive config resolution (Flag > Env > Prompt > Default), and graceful signal handling. No global state. No external dependencies — stdlib only.

## Packages

### `logger`

Colored, level-aware output. All methods accept a printf-style format string and variadic args.

| Method | Output | Destination | Suppressed by quiet? |
|--------|--------|-------------|----------------------|
| `Print(msg, v...)` | plain, no prefix | stdout | yes |
| `Info(msg, v...)` | cyan `[INFO]` | stdout | yes |
| `Success(msg, v...)` | green `[SUCCESS]` | stdout | yes |
| `Detail(msg, v...)` | plain `-------` | stdout (verbose only) | yes |
| `Debug(msg, v...)` | red `[--dEbUg--]` | stdout | no |
| `Warn(msg, v...)` | yellow `[WARNING]` | stderr | no |
| `Error(msg, v...)` | red `[ERROR]` | stderr | no |
| `Fatal(msg, v...)` | magenta `[FATAL]` | stderr, then `os.Exit(1)` | no |

**Verbose mode:** `logger.New(true)` enables `Detail` output.

**Quiet mode:** `.WithQuiet()` suppresses `Print`, `Info`, `Success`, and `Detail`. Use for `--quiet` flags. `Warn`, `Error`, `Fatal`, and `Debug` are always shown.

**Custom writers:** `.WithWriters(out, err)` replaces `os.Stdout`/`os.Stderr`. Useful for testing or writing to a file.

**`Debug`** always prints regardless of verbose or quiet mode — use it during development and remove it before shipping. Use `Detail` for output that should only appear with `--verbose`.

### `ui`

Interactive prompts, TTY detection, config resolution, and signal handling. All in one session object.

**`ResolveString`** applies precedence: Flag > Env > Prompt > Default. Pass the flag's current value and whether it was explicitly set — works with any flag library (cobra, pflag, stdlib `flag`, etc).

**`ResolveBool`** applies precedence: Flag > Env > Default. There is no prompt tier for booleans — use `Confirm` for interactive boolean input. Env values of `"true"`, `"1"`, or `"yes"` (case-insensitive) are treated as true.

**`Confirm`** prompts for y/n and returns `true` without prompting when not interactive, so automation and CI pipelines proceed unblocked.

**Signal handling** via `WithInterrupt` registers SIGINT/SIGTERM handlers and stores a cancellable context on `ui.Ctx`. Pass `ui.Ctx` to long-running operations so Ctrl+C cancels them cleanly. Call `StopSignal()` (typically via `defer`) to release the handler when the command exits. `ui.Ctx` is always safe to use — it is `context.Background()` until `WithInterrupt` is called.

### `sys`

`ContextWithInterrupt(parent)` returns a context that cancels on SIGINT or SIGTERM. Use this when you need signal-aware cancellation without a `UI` — for example in background workers or daemons. For interactive CLI tools, prefer `ui.WithInterrupt`.

## Usage

```go
func execute(cmd *cobra.Command, args []string) {
    // 1. Init (no globals)
    log := logger.New(verbose).WithQuiet()  // omit WithQuiet() if no --quiet flag
    userInterface := ui.New(nonInteractive).
        WithLogger(log).
        WithInterrupt(context.Background())
    defer userInterface.StopSignal()

    // 2. Resolve config: flag > env > prompt > default
    // Works with any flag library — just pass the value and whether it was set
    inputVal, _ := cmd.Flags().GetString("input")
    userInterface.ResolveString(inputVal, cmd.Flags().Changed("input"), "TOOL_INPUT", "Enter URL", &cfg.Input)

    dryRunVal, _ := cmd.Flags().GetBool("dry-run")
    userInterface.ResolveBool(dryRunVal, cmd.Flags().Changed("dry-run"), "TOOL_DRY_RUN", &cfg.DryRun)

    // 3. Summarize & confirm
    log.Print("Target: %s", cfg.Input)
    if !userInterface.Confirm("Proceed?") {
        return
    }

    // 4. Pass ui.Ctx to long-running work so Ctrl+C cancels it
    log.Info("Working...")
    doWork(userInterface.Ctx, cfg)
}
```

Tools using cmdkit should support a `--dry-run` flag and use the logger to report what would have been done instead of performing destructive actions.
