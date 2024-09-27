import { CopySimple } from '~/console/components/icons';
import hljs from 'highlight.js';
import { useEffect, useRef } from 'react';
import { toast } from '~/components/molecule/toast';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { cn } from '~/components/utils';

interface ICodeView {
  data: string;
  copy: boolean;
  showShellPrompt?: boolean;
  language?: string;
  title?: string;
  preClassName?: string;
  isMultilineData?: boolean;
}
const CodeView = ({
  data,
  copy,
  showShellPrompt: _,
  language = 'shell',
  title,
  preClassName,
  isMultilineData,
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
        {isMultilineData ? (
          <div
            onClick={() => {
              if (copy) cpy(data);
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                if (copy) cpy(data);
              }
            }}
            className="group/sha cursor-pointer hljs p-lg relative"
          >
            <pre className={cn('flex-1 overflow-auto', preClassName)}>
              <code ref={ref}>{data}</code>
            </pre>
            <span className="absolute mr-2xl mt-2xl top-0 right-0">
              <CopySimple size={14} />
            </span>
          </div>
        ) : (
          <div
            onClick={() => {
              if (copy) cpy(data);
            }}
            className="group/sha cursor-pointer p-lg rounded-md bodyMd flex flex-row gap-xl items-center hljs w-full"
          >
            <pre className={cn('flex-1 overflow-auto', preClassName)}>
              <code ref={ref}>{data}</code>
            </pre>
            <span className="invisible group-hover/sha:visible">
              <CopySimple size={14} />
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

export default CodeView;
