function postMessage(message: string) {
  fetch('http://localhost:8080/message', {
    method: 'POST',
    body: message,
  })
    .then((response) => response.text())
    .then((data) => console.log(data))
    .catch((error) => console.error('Error:', error));
}

function dumpConsoleToFile() {
  const { log } = console;
  const { error } = console;
  const { warn } = console;
  const { info } = console;

  console.log = (...args) => {
    log.apply(console, args);
    postMessage(
      `LOG: ${args.map((a) => JSON.stringify(a, null, 2)).join(' ')}\n`
    );
  };

  console.error = (...args) => {
    error.apply(console, args);
    postMessage(`ERROR: ${args.join(' ')}\n`);
  };

  console.warn = (...args) => {
    warn.apply(console, args);
    postMessage(`WARN: ${args.join(' ')}\n`);
  };

  console.info = (...args) => {
    info.apply(console, args);
    postMessage(`INFO: ${args.join(' ')}\n`);
  };
}

dumpConsoleToFile();
