import CodeMirrorClient from '~/root/lib/client/components/editor-client';

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
