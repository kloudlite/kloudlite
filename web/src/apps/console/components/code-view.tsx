import { CopySimple } from '@jengaicons/react';
import { toast } from '~/components/molecule/toast';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';

interface ICodeView {
  data: string;
  copy: boolean;
  showShellPrompt?: boolean;
}
const CodeView = ({ data, copy, showShellPrompt }: ICodeView) => {
  const { copy: cpy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });
  return (
    <div
      onClick={() => {
        if (copy) cpy(data);
      }}
      className="group/sha cursor-pointer bg-surface-basic-active p-lg rounded-lg bodyMd flex flex-row gap-xl items-center"
    >
      <code className="flex flex-row items-center gap-lg flex-1 break-all">
        <span className="opacity-60">{showShellPrompt && '$'}</span>
        <div>{data}</div>
      </code>
      <span className="invisible group-hover/sha:visible">
        <CopySimple size={14} />
      </span>
    </div>
  );
};

export default CodeView;
