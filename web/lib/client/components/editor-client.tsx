import { Editor, EditorProps } from '@monaco-editor/react';
import { cn } from '~/components/utils';

type codelang = 'yaml' | 'json';
// @ts-ignore

const CodeEditorClient = (
  props: EditorProps & {
    lang?: codelang;
  }
) => {
  const { lang, className } = props;
  return (
    <Editor
      // className="h-full w-full border-text-soft rounded-sm"
      // height="90vh"
      {...{
        // theme: 'vs-dark',
        defaultLanguage: lang,
        onValidate: (v) => {
          console.log(v);
        },
        options: {
          padding: {
            top: 20,
            bottom: 20,
          },
          tabSize: 2,
          fontSize: 18,
          minimap: {
            enabled: false,
          },
        },
        ...props,
        className: cn(
          className,
          'h-full w-full border border-text-soft rounded-sm'
        ),
      }}
    />
  );
};

export default CodeEditorClient;
