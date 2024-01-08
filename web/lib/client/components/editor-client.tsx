import { ReactCodeMirrorProps } from '@uiw/react-codemirror';
import React, { Suspense } from 'react';

type codelang = 'yaml' | 'json';
// @ts-ignore
const CodeMirrorMain = React.lazy(() => import('./editor'));

const CodeMirrorClient = (
  props: ReactCodeMirrorProps & {
    lang?: codelang;
  }
) => {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <CodeMirrorMain {...props} />
    </Suspense>
  );
};

export default CodeMirrorClient;
