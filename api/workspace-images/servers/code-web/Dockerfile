FROM kloudlite/workspace-base:latest

USER root

# Install Node.js and npm
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

# Install serve for serving static files
RUN npm install -g serve

USER workspace

# Create a simple Monaco editor setup
RUN mkdir -p /home/workspace/code-web && \
    cd /home/workspace/code-web && \
    cat > index.html <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Code Web - Monaco Editor</title>
    <link rel="stylesheet" data-name="vs/editor/editor.main" href="https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs/editor/editor.main.css">
</head>
<body style="margin: 0; padding: 0; height: 100vh;">
    <div id="container" style="width: 100%; height: 100%;"></div>
    <script src="https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs/loader.js"></script>
    <script>
        require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' }});
        require(['vs/editor/editor.main'], function() {
            monaco.editor.create(document.getElementById('container'), {
                value: '// Welcome to Code Web\\n// Start coding!\\n',
                language: 'javascript',
                theme: 'vs-dark'
            });
        });
    </script>
</body>
</html>
EOF

EXPOSE 3000

CMD ["serve", "-s", "/home/workspace/code-web", "-l", "3000"]
