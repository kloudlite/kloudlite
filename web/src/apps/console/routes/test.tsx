import { useState } from 'react';
import CodeEditorClient from '~/root/lib/client/components/editor-client';

function App() {
  const [code, setCode] = useState(`
name: hello
age: 30
`);

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1">
        <CodeEditorClient
          className="flex-1"
          lang="yaml"
          onChange={(v) => {
            setCode(v || '');
          }}
          value={`
name: hello
age: 30
`}
        />
      </div>

      <pre className="px-xl flex-1">{code}</pre>
    </div>
  );
}

export default App;
