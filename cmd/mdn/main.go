package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/app"
	"md-notes/internal/buffer"
	"md-notes/internal/watcher"
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

	// Setup file watcher for non-scratchpad, non-pipe files
	var w *watcher.Watcher
	if !buf.IsStdinBuffer && buf.FilePath != "" {
		w, err = watcher.NewWatcher()
		if err != nil {
			w = nil
		}
	}

	model := app.NewModel(buf, w)

	var programOpts []tea.ProgramOption
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
