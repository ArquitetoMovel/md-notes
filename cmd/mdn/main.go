package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/app"
	"md-notes/internal/buffer"
	"md-notes/internal/config"
	"md-notes/internal/theme"
	"md-notes/internal/watcher"
)

var (
	// Version is the current version, injected during build via ldflags.
	Version = "dev"
	// GitCommit is the git SHA, injected during build via ldflags.
	GitCommit = "none"
	// BuildDate is the build timestamp, injected during build via ldflags.
	BuildDate = "unknown"
)

func openTTY() (io.Reader, io.Closer, error) {
	ttyDevice := "/dev/tty"
	if runtime.GOOS == "windows" {
		ttyDevice = "CONIN$"
	}

	tty, err := os.Open(ttyDevice)
	if err != nil {
		return nil, nil, err
	}
	return tty, tty, nil
}

// Execute runs the CLI entrypoint with configurable I/O streams and optional Bubble Tea program options.
func Execute(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, customOpts ...tea.ProgramOption) int {
	var buf *buffer.Buffer
	var err error
	var isPipe bool

	if stdinFile, ok := stdin.(*os.File); ok {
		stat, statErr := stdinFile.Stat()
		if statErr == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			isPipe = true
		}
	} else if stdin != nil {
		isPipe = true
	}

	if len(args) > 1 {
		switch args[1] {
		case "-v", "--version", "version":
			fmt.Fprintf(stdout, "mdn version %s (%s, %s, %s/%s)\n", Version, GitCommit, BuildDate, runtime.GOOS, runtime.GOARCH)
			return 0
		case "-h", "--help", "help":
			fmt.Fprintf(stdout, "mdn - Markdown Editor & Viewer for Terminal\n\n"+
				"Usage:\n"+
				"  mdn [file.md]       Open or edit markdown file\n"+
				"  mdn                 Open scratchpad buffer\n"+
				"  cat file | mdn      Read from stdin pipe\n\n"+
				"Flags:\n"+
				"  -v, --version       Show version information\n"+
				"  -h, --help          Show help information\n")
			return 0
		}
	}

	if isPipe {
		buf, err = buffer.LoadFromReader(stdin, "[Stdin Buffer]")
		if err != nil {
			fmt.Fprintf(stderr, "Erro: Falha ao ler stream de entrada: %v\n", err)
			return 1
		}
	} else if len(args) > 1 {
		filePath := args[1]
		buf, err = buffer.LoadFromFile(filePath)
		if err != nil {
			fmt.Fprintf(stderr, "%s\n", err.Error())
			return 1
		}
	} else {
		buf = buffer.NewScratchpadBuffer()
	}

	// Load configuration and initialize compiled theme
	cfg, warnings, _ := config.LoadConfig()
	resolvedTheme := theme.ResolveThemeName(cfg.Theme)
	palette, _ := theme.GetPalette(resolvedTheme)
	compiledTheme := theme.CompileTheme(palette, cfg.Colors)

	// Setup file watcher for non-scratchpad, non-pipe files
	var w *watcher.Watcher
	if !buf.IsStdinBuffer && buf.FilePath != "" {
		w, err = watcher.NewWatcher()
		if err != nil {
			w = nil
		}
	}

	// Setup configuration file watcher for dynamic hot-reload
	var configWatcher *config.Watcher
	configPath, configPathErr := config.GetConfigFilePath()
	if configPathErr == nil {
		configWatcher, err = config.NewWatcher()
		if err != nil {
			configWatcher = nil
		}
	}

	model := app.NewModel(
		buf,
		w,
		app.WithConfig(cfg),
		app.WithTheme(compiledTheme),
		app.WithConfigWatcher(configWatcher),
	)
	if len(warnings) > 0 {
		model.StatusMsg = strings.Join(warnings, " | ")
	}

	var programOpts []tea.ProgramOption
	programOpts = append(programOpts, tea.WithAltScreen(), tea.WithMouseCellMotion())
	var ttyCloser io.Closer

	if isPipe {
		if _, isFile := stdin.(*os.File); isFile {
			tty, closer, ttyErr := openTTY()
			if ttyErr == nil {
				programOpts = append(programOpts, tea.WithInput(tty))
				ttyCloser = closer
			}
		}
	}

	if stdout != nil {
		programOpts = append(programOpts, tea.WithOutput(stdout))
	}
	programOpts = append(programOpts, customOpts...)

	p := tea.NewProgram(model, programOpts...)

	if w != nil && buf.FilePath != "" {
		_ = w.Start(buf.FilePath, p.Send)
		defer w.Close()
	}

	if configWatcher != nil && configPath != "" {
		_ = configWatcher.Start(configPath, p.Send)
		defer configWatcher.Close()
	}

	if ttyCloser != nil {
		defer ttyCloser.Close()
	}

	if _, runErr := p.Run(); runErr != nil {
		fmt.Fprintf(stderr, "Erro ao executar interface: %v\n", runErr)
		return 1
	}

	return 0
}

func main() {
	os.Exit(Execute(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
