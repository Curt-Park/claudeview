package transcript

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchEvent signals a new entry in a watched file.
type WatchEvent struct {
	FilePath string
	Entry    Entry
}

// Watcher watches multiple JSONL files and emits new entries.
type Watcher struct {
	fsw    *fsnotify.Watcher
	events chan WatchEvent
	errors chan error
	files  map[string]*fileState
	mu     sync.Mutex
	done   chan struct{}
}

type fileState struct {
	path   string
	offset int64
}

// NewWatcher creates a new multi-file watcher.
func NewWatcher() (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		fsw:    fsw,
		events: make(chan WatchEvent, 256),
		errors: make(chan error, 16),
		files:  make(map[string]*fileState),
		done:   make(chan struct{}),
	}
	go w.run()
	return w, nil
}

// Watch adds a file to be watched. It reads existing content first.
func (w *Watcher) Watch(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.files[path]; ok {
		return nil // already watching
	}

	offset, err := w.readAll(path, 0)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	w.files[path] = &fileState{path: path, offset: offset}
	return w.fsw.Add(path)
}

// Events returns the channel for receiving new entries.
func (w *Watcher) Events() <-chan WatchEvent {
	return w.events
}

// Errors returns the error channel.
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsw.Close()
}

func (w *Watcher) run() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-w.done:
			return
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if ev.Has(fsnotify.Write) || ev.Has(fsnotify.Create) {
				w.readNew(ev.Name)
			}
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			select {
			case w.errors <- err:
			default:
			}
		case <-ticker.C:
			// Poll for any missed updates
			w.mu.Lock()
			paths := make([]string, 0, len(w.files))
			for p := range w.files {
				paths = append(paths, p)
			}
			w.mu.Unlock()
			for _, p := range paths {
				w.readNew(p)
			}
		}
	}
}

func (w *Watcher) readNew(path string) {
	w.mu.Lock()
	state, ok := w.files[path]
	if !ok {
		w.mu.Unlock()
		return
	}
	offset := state.offset
	w.mu.Unlock()

	newOffset, err := w.readAll(path, offset)
	if err != nil {
		return
	}

	w.mu.Lock()
	if s, ok := w.files[path]; ok {
		s.offset = newOffset
	}
	w.mu.Unlock()
}

func (w *Watcher) readAll(path string, fromOffset int64) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return fromOffset, err
	}
	defer f.Close()

	if fromOffset > 0 {
		if _, err := f.Seek(fromOffset, io.SeekStart); err != nil {
			return fromOffset, err
		}
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	currentOffset := fromOffset

	for scanner.Scan() {
		line := scanner.Bytes()
		currentOffset += int64(len(line)) + 1 // +1 for newline

		if len(line) == 0 {
			continue
		}
		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		select {
		case w.events <- WatchEvent{FilePath: path, Entry: entry}:
		case <-w.done:
			return currentOffset, nil
		}
	}

	return currentOffset, scanner.Err()
}
