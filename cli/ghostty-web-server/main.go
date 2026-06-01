package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

const html = `<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>terminal</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body, #terminal { height: 100%; width: 100%; overflow: hidden; }
body { background: #161b22; }
</style>
</head>
<body>
<div id="terminal"></div>
<script type="module">
import { init, Terminal, FitAddon } from '/dist/ghostty-web.js';
await init();
const term = new Terminal({
  cols: 80, rows: 24,
  fontSize: 14,
  fontFamily: 'JetBrainsMono Nerd Font Mono, FiraCode Nerd Font Mono, Menlo, monospace',
  theme: {
    background: '#161b22',
    foreground: '#c9d1d9',
    cursor: '#58a6ff',
    selectionBackground: '#264f78',
    selectionForeground: '#c9d1d9',
    black: '#21262d', red: '#f85149', green: '#3fb950', yellow: '#d29922',
    blue: '#58a6ff', magenta: '#bc8cff', cyan: '#39d2c0', white: '#c9d1d9',
    brightBlack: '#6e7681', brightRed: '#ff7b72', brightGreen: '#56d364',
    brightYellow: '#e3b341', brightBlue: '#79c0ff', brightMagenta: '#d2a8ff',
    brightCyan: '#56d4dd', brightWhite: '#f0f6fc',
  },
});

const fitAddon = new FitAddon();
term.loadAddon(fitAddon);
const container = document.getElementById('terminal');
await term.open(container);
fitAddon.fit();
window.addEventListener('resize', () => fitAddon.fit());

const ws = new WebSocket(((location.protocol === 'https:') ? 'wss://' : 'ws://') + location.host + '/ws?cols=' + term.cols + '&rows=' + term.rows);

ws.onopen = () => { term.focus(); };
ws.onmessage = (e) => {
  if (typeof e.data === 'string' && e.data.startsWith('{')) {
    try {
      const msg = JSON.parse(e.data);
      if (msg.type === 'clipboard' && msg.data) {
        // Buffer clipboard data — browser requires user gesture for clipboard write.
        // Flush on next keydown or click event.
        pendingClipboard = msg.data;
        return;
      }
    } catch (_) {}
  }
  term.write(e.data);
};
ws.onclose = () => { term.write('\r\n\x1b[31mConnection closed. Reconnecting...\x1b[0m\r\n'); setTimeout(() => location.reload(), 2000); };
term.onData((data) => { if (ws.readyState === WebSocket.OPEN) ws.send(data); });
term.onResize(({cols, rows}) => { if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({type:'resize',cols,rows})); });

// Clipboard support
// Intercept Ctrl+C when text is selected to copy to browser clipboard
// Also: buffer OSC-52 clipboard data and flush on next user gesture
let pendingClipboard = '';
const flushClipboard = () => {
  if (!pendingClipboard) return;
  const t = pendingClipboard;
  pendingClipboard = '';
  const copy = (v) => {
    const ta = document.createElement('textarea');
    ta.value = v; ta.style.position = 'fixed'; ta.style.left = '-9999px';
    document.body.appendChild(ta); ta.select();
    try { document.execCommand('copy'); } catch (_) {}
    document.body.removeChild(ta);
  };
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(t).catch(() => copy(t));
  } else {
    copy(t);
  }
};
document.addEventListener('keydown', flushClipboard, true);
document.addEventListener('click', flushClipboard);

document.addEventListener('keydown', (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'c' && !e.shiftKey) {
    const sel = term.getSelection();
    if (!sel) return; // No selection, let Ctrl+C pass through to PTY (SIGINT)
    e.preventDefault();
    e.stopImmediatePropagation();
    const copyViaExec = (t) => {
      const ta = document.createElement('textarea');
      ta.value = t; ta.style.position = 'fixed'; ta.style.left = '-9999px';
      document.body.appendChild(ta); ta.select();
      try { document.execCommand('copy'); } catch (_) {}
      document.body.removeChild(ta);
    };
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(sel).catch(() => copyViaExec(sel));
    } else {
      copyViaExec(sel);
    }
  }
}, true); // Capturing phase — fires before terminal canvas

// Paste: send clipboard text to terminal
document.addEventListener('paste', async (e) => {
  const text = e.clipboardData.getData('text');
  if (text && ws.readyState === WebSocket.OPEN) {
    ws.send(text);
  }
});

// Copy: Ctrl+Shift+C to copy selection to clipboard
document.addEventListener('keydown', async (e) => {
  if (e.key === 'c' && e.ctrlKey && e.shiftKey && selection) {
    e.preventDefault();
    try {
      await navigator.clipboard.writeText(selection);
    } catch (_) {
      // Fallback: use execCommand
      const ta = document.createElement('textarea');
      ta.value = selection;
      ta.style.position = 'fixed'; ta.style.opacity = '0';
      document.body.appendChild(ta); ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
    }
  }
});

// Click to focus terminal
container.addEventListener('click', () => term.focus());
</script>
</body>
</html>`

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type resizeMsg struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

type clipboardMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

var osc52Regex = regexp.MustCompile(`\x1b\]52;(?:[^;]*);([a-zA-Z0-9+/=]*)(?:\x07|\x1b\\)`)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	shellCmd := os.Getenv("SHELL")
	if shellCmd == "" {
		shellCmd = "/bin/bash"
	}
	assetDir := os.Getenv("GHOSTTY_WEB_ASSETS")
	if assetDir == "" {
		assetDir = "/usr/local/share/ghostty-web"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir(assetDir))))
	mux.HandleFunc("/ghostty-vt.wasm", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, assetDir+"/ghostty-vt.wasm")
	})
	mux.HandleFunc("/ws", handleWS(shellCmd))

	server := &http.Server{Addr: ":" + port, Handler: mux}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Fprintf(os.Stderr, "ghostty-web listening on :%s shell=%s\n", port, shellCmd)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "ghostty-web: %v\n", err)
		}
	}()

	ctx := context.Background()
	select {
	case <-quit:
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
}

func handleWS(shellCmd string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		cols, _ := strconv.Atoi(r.URL.Query().Get("cols"))
		rows, _ := strconv.Atoi(r.URL.Query().Get("rows"))
		if cols == 0 {
			cols = 80
		}
		if rows == 0 {
			rows = 24
		}

		shell := strings.Fields(shellCmd)
		cmd := exec.Command(shell[0], shell[1:]...)
		cmd.Env = append(os.Environ(), "TERM=xterm-256color", "COLORTERM=truecolor")
		if home := os.Getenv("HOME"); home != "" {
			cmd.Dir = home
		} else {
			cmd.Dir = "/"
		}

		f, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\x1b[31m%s\x1b[0m\r\n", err)))
			return
		}
		defer f.Close()

		done := make(chan struct{}, 1)

		go func() {
			buf := make([]byte, 4096)
			writer := &wsWriter{conn: conn}
			for {
				n, err := f.Read(buf)
				if err != nil {
					break
				}
				data := buf[:n]
				// Intercept OSC 52 clipboard sequences
				if idx := osc52Regex.FindSubmatchIndex(data); idx != nil {
					// Send everything before the OSC 52 sequence
					writer.Write(data[:idx[0]])
					// Decode base64 clipboard data
					b64 := data[idx[2]:idx[3]]
					decoded, decErr := base64.StdEncoding.DecodeString(string(b64))
					if decErr == nil {
						msg, _ := json.Marshal(clipboardMsg{Type: "clipboard", Data: string(decoded)})
						conn.WriteMessage(websocket.TextMessage, msg)
					}
					// Send the rest after the OSC 52 sequence
					end := idx[1]
					writer.Write(data[end:])
				} else {
					writer.Write(data)
				}
			}
			done <- struct{}{}
		}()

		go func() {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					break
				}
				if len(msg) > 0 && msg[0] == '{' {
					var resize resizeMsg
					if json.Unmarshal(msg, &resize) == nil && resize.Type == "resize" {
						pty.Setsize(f, &pty.Winsize{Rows: resize.Rows, Cols: resize.Cols})
						continue
					}
				}
				f.Write(msg)
			}
			done <- struct{}{}
		}()

		<-done
		cmd.Process.Signal(syscall.SIGTERM)
		cmd.Wait()
	}
}

type wsWriter struct {
	conn *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
