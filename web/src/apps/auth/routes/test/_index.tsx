import CodeMirrorClient from './editor-client';

const Editor = () => {
  return (
    <div>
      <h1>Editor</h1>
      <CodeMirrorClient
        value={`
---
title: Hello World
---

`}
      />
    </div>
  );
};

export default Editor;
