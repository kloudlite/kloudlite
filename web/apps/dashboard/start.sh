#!/bin/sh

# Start Next.js server on internal port (background)
PORT=3002 bun /app/apps/dashboard/server.js &

# Start proxy server on main port (foreground)
exec bun /app/apps/dashboard/proxy-server.js
