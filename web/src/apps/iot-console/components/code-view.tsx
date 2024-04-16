import { CopySimple } from '@jengaicons/react';
import hljs from 'highlight.js';
import { useEffect, useRef } from 'react';
import { toast } from '~/components/molecule/toast';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';

interface ICodeView {
  data: string;
  copy: boolean;
  showShellPrompt?: boolean;
  language?: string;
  title?: string;
}
const CodeView = ({
  data,
  copy,
  showShellPrompt: _,
  language = 'shell',
  title,
}: ICodeView) => {
  const { copy: cpy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const ref = useRef(null);

  useEffect(() => {
    (async () => {
      if (ref.current) {
        const hr = hljs.highlight(
          data,
          {
            language,
          },
          false
        );

        // @ts-ignore
        ref.current.innerHTML = hr.value;
      }
    })();
  }, [data, language]);

  return (
    <div className="flex flex-col gap-lg flex-1 min-w-[45%]">
      {!!title && (
        <div className="bodyMd-medium text-text-default">{title}</div>
      )}
      <div className="bodyMd text-text-strong">
        <div
          onClick={() => {
            if (copy) cpy(data);
          }}
          className="group/sha cursor-pointer p-lg rounded-md bodyMd flex flex-row gap-xl items-center hljs w-full"
        >
          <pre className="flex-1 overflow-auto">
            <code ref={ref}>{data}</code>
          </pre>

          <span className="invisible group-hover/sha:visible">
            <CopySimple size={14} />
          </span>
        </div>
      </div>
    </div>
  );
};

export default CodeView;
