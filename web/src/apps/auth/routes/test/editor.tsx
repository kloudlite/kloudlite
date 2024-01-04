import CodeM, { ReactCodeMirrorProps } from '@uiw/react-codemirror';
import { StreamLanguage } from '@codemirror/language';
import { yaml } from '@codemirror/legacy-modes/mode/yaml';
import { json } from '@codemirror/legacy-modes/mode/javascript';

type codelang = 'yaml' | 'json';

const getStreamLanguage = (lang: codelang) => {
  switch (lang) {
    case 'yaml':
      return StreamLanguage.define(yaml);
    case 'json':
      return StreamLanguage.define(json);
    default:
      return StreamLanguage.define(yaml);
  }
};

const CodeMirror = (
  props: ReactCodeMirrorProps & {
    lang?: codelang;
  }
) => {
  const { lang } = props;

  return (
    <CodeM
      {...{
        ...props,
        ...{
          extensions: [getStreamLanguage(lang || 'yaml')],
        },
      }}
    />
  );
};

export default CodeMirror;
