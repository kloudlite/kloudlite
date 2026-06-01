package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
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
body { background: #1e1e1e; overflow: hidden; }
#terminal { width: 100vw; height: 100vh; }
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
  fontFamily: 'JetBrainsMono Nerd Font Mono, FiraCode Nerd Font Mono, monospace',
  theme: { background: '#1e1e1e', foreground: '#d4d4d4' },
});
const fitAddon = new FitAddon();
term.loadAddon(fitAddon);
await term.open(document.getElementById('terminal'));
fitAddon.fit();

const ws = new WebSocket(((location.protocol === 'https:') ? 'wss://' : 'ws://') + location.host + '/ws?cols=' + term.cols + '&rows=' + term.rows);

ws.onopen = () => { term.focus(); };
ws.onmessage = (e) => { term.write(e.data); };
ws.onclose = () => { term.write('\r\n\x1b[31mConnection closed.\x1b[0m\r\n'); };
term.onData((data) => { if (ws.readyState === WebSocket.OPEN) ws.send(data); });
term.onResize(({cols, rows}) => { if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({type:'resize',cols,rows})); });
window.addEventListener('resize', () => fitAddon.fit());
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
			io.Copy(&wsWriter{conn: conn}, f)
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
