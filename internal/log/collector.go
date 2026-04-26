package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Collector はサービスのstdout/stderrをJSONLファイルに収集する
// 実プロセス用には *os.File、スクリプト用には io.Writer を返す。
type Collector struct {
	logDir   string
	service  string
	worktree string
	file     *os.File
	mu       sync.Mutex
	wg       sync.WaitGroup

	maxSize  int64
	maxFiles int

	stdoutPipeR *os.File
	stdoutPipeW *os.File
	stderrPipeR *os.File
	stderrPipeW *os.File
}

// NewCollector はログコレクターを生成する
func NewCollector(sangoDir, worktree, service string) (*Collector, error) {
	logDir := filepath.Join(sangoDir, "logs", worktree)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	logPath := filepath.Join(logDir, service+".jsonl")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &Collector{
		logDir:   logDir,
		service:  service,
		worktree: worktree,
		file:     f,
		maxSize:  50 * 1024 * 1024,
		maxFiles: 5,
	}, nil
}

// StdoutFile は stdout 用の *os.File を返す。
func (c *Collector) StdoutFile() (*os.File, error) {
	return c.ensurePipe("stdout")
}

// StderrFile は stderr 用の *os.File を返す。
func (c *Collector) StderrFile() (*os.File, error) {
	return c.ensurePipe("stderr")
}

// StdoutWriter は stdout 用 writer を返す。
func (c *Collector) StdoutWriter() (io.Writer, error) {
	return c.StdoutFile()
}

// StderrWriter は stderr 用 writer を返す。
func (c *Collector) StderrWriter() (io.Writer, error) {
	return c.StderrFile()
}

func (c *Collector) ensurePipe(stream string) (*os.File, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch stream {
	case "stdout":
		if c.stdoutPipeW != nil {
			return c.stdoutPipeW, nil
		}
		r, w, err := os.Pipe()
		if err != nil {
			return nil, err
		}
		c.stdoutPipeR = r
		c.stdoutPipeW = w
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.scanAndWrite(r, "stdout", os.Stdout)
			_ = r.Close()
		}()
		return w, nil
	case "stderr":
		if c.stderrPipeW != nil {
			return c.stderrPipeW, nil
		}
		r, w, err := os.Pipe()
		if err != nil {
			return nil, err
		}
		c.stderrPipeR = r
		c.stderrPipeW = w
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.scanAndWrite(r, "stderr", os.Stderr)
			_ = r.Close()
		}()
		return w, nil
	default:
		return nil, fmt.Errorf("unsupported stream: %s", stream)
	}
}

// scanAndWrite はパイプから行を読み取り、ターミナルとJSONLファイルの両方に書き込む
func (c *Collector) scanAndWrite(r io.Reader, stream string, terminal *os.File) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		_, _ = terminal.WriteString(line + "\n")

		entry := &LogEntry{
			Timestamp: time.Now(),
			Service:   c.service,
			Worktree:  c.worktree,
			Stream:    stream,
			Level:     DetectLevel(line, stream),
			Message:   line,
		}

		data, err := entry.Marshal()
		if err != nil {
			fallback := fmt.Sprintf(`{"ts":"%s","svc":"%s","wt":"%s","stream":"%s","msg":"[marshal error] %s"}`,
				time.Now().Format(time.RFC3339), c.service, c.worktree, stream, line)
			data = []byte(fallback)
		}

		c.mu.Lock()
		c.writeWithRotation(append(data, '\n'))
		c.mu.Unlock()
	}
}

// writeWithRotation はサイズチェック付きでファイルに書き込む
func (c *Collector) writeWithRotation(data []byte) {
	info, err := c.file.Stat()
	if err == nil && info.Size()+int64(len(data)) > c.maxSize {
		c.rotate()
	}
	_, _ = c.file.Write(data)
}

// rotate はログファイルをローテーションする
func (c *Collector) rotate() {
	basePath := filepath.Join(c.logDir, c.service+".jsonl")
	_ = c.file.Close()
	Rotate(basePath, c.maxFiles)

	f, err := os.OpenFile(basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		f, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	}
	c.file = f
}

// SetRotationConfig はローテーション設定を変更する
func (c *Collector) SetRotationConfig(maxSize int64, maxFiles int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxSize = maxSize
	c.maxFiles = maxFiles
}

// Close はログファイルとパイプを閉じる
func (c *Collector) Close() error {
	c.mu.Lock()
	stdoutW := c.stdoutPipeW
	stderrW := c.stderrPipeW
	c.stdoutPipeW = nil
	c.stderrPipeW = nil
	c.mu.Unlock()

	if stdoutW != nil {
		_ = stdoutW.Close()
	}
	if stderrW != nil {
		_ = stderrW.Close()
	}
	c.wg.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.file.Close()
}
