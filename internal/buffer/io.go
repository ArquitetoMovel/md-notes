package buffer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrNoFileName = errors.New("Erro: Nenhum nome de arquivo definido. Use :w <caminho>")
)

// LoadFromFile reads an existing file or initializes a new buffer if the file does not exist.
func LoadFromFile(path string) (*Buffer, error) {
	if path == "" {
		return NewScratchpadBuffer(), nil
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewEmptyBuffer(path), nil
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("Erro: Permissão negada ao abrir '%s'", path)
		}
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("'%s' é um diretório", path)
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("Erro: Permissão negada ao abrir '%s'", path)
		}
		return nil, err
	}
	defer file.Close()

	buf, err := LoadFromReader(file, path)
	if err != nil {
		return nil, err
	}

	buf.IsNewFile = false
	buf.IsStdinBuffer = false
	buf.OriginalModTime = info.ModTime()
	buf.OriginalMode = info.Mode()
	return buf, nil
}

// LoadFromReader reads entire content from an io.Reader and creates a buffer.
func LoadFromReader(r io.Reader, name string) (*Buffer, error) {
	var lines []Line
	reader := bufio.NewReader(r)

	for {
		lineBytes, err := reader.ReadBytes('\n')
		if len(lineBytes) > 0 {
			ending := EndingNone
			if bytes.HasSuffix(lineBytes, []byte("\r\n")) {
				ending = EndingCRLF
			} else if bytes.HasSuffix(lineBytes, []byte("\n")) {
				ending = EndingLF
			}

			content := string(lineBytes)
			lines = append(lines, NewLine(content, ending))
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}

	// If stream was empty, provide at least one empty line
	if len(lines) == 0 {
		lines = []Line{NewLine("", EndingLF)}
	}

	isStdin := name == "[Stdin Buffer]"
	return &Buffer{
		Lines:           lines,
		Cursor:          Cursor{Line: 0, Col: 0},
		FilePath:        name,
		IsDirty:         false,
		IsNewFile:       !isStdin && name == "",
		IsStdinBuffer:   isStdin,
		OriginalModTime: time.Now(),
		OriginalMode:    0644,
	}, nil
}

// AtomicSave securely writes the buffer to disk using a temporary file and rename.
func AtomicSave(buf *Buffer, targetPath string) error {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	savePath := targetPath
	if savePath == "" {
		savePath = buf.FilePath
	}

	if savePath == "" || savePath == "[Stdin Buffer]" {
		return ErrNoFileName
	}

	// Create parent directories if they do not exist
	dir := filepath.Dir(savePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Erro: Não foi possível gravar o arquivo em '%s': %w", savePath, err)
		}
	}

	// Create temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, ".tmp.mdn-*")
	if err != nil {
		return fmt.Errorf("Erro: Não foi possível gravar o arquivo em '%s': %w", savePath, err)
	}
	tempPath := tempFile.Name()

	// Ensure cleanup on failure
	cleanup := true
	defer func() {
		if cleanup {
			tempFile.Close()
			os.Remove(tempPath)
		}
	}()

	writer := bufio.NewWriter(tempFile)
	for _, line := range buf.Lines {
		if _, err := writer.WriteString(line.RawString()); err != nil {
			return fmt.Errorf("Erro: Não foi possível gravar o arquivo em '%s': %w", savePath, err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Erro: Não foi possível gravar o arquivo em '%s': %w", savePath, err)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("Erro: Não foi possível gravar o arquivo em '%s': %w", savePath, err)
	}

	// Preserve file permissions
	mode := buf.OriginalMode
	if mode == 0 {
		mode = 0644
	}
	if err := tempFile.Chmod(mode); err != nil {
		// Non-fatal on some OS/filesystems, continue
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("Erro: Não foi possível gravar o arquivo em '%s': %w", savePath, err)
	}

	// Atomic replace
	if err := os.Rename(tempPath, savePath); err != nil {
		return fmt.Errorf("Erro: Não foi possível gravar o arquivo em '%s': %w", savePath, err)
	}

	cleanup = false
	buf.FilePath = savePath
	buf.IsNewFile = false
	buf.IsDirty = false
	buf.IsStdinBuffer = false
	buf.OriginalModTime = time.Now()

	return nil
}
