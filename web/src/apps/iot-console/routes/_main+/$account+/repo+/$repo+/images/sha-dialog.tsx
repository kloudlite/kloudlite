import { CodeSimpleFill, Tag, TerminalWindow } from '~/iotconsole/components/icons';
import { useParams } from '@remix-run/react';
import Popup from '~/components/molecule/popup';
import CodeView from '~/iotconsole/components/code-view';
import { IDialog } from '~/iotconsole/components/types.d';
import { registryHost } from '~/root/lib/configs/base-url.cjs';
import { useOutletContext } from 'react-router-dom';
import { ISHADialogData } from './tags-resources';
import { IRepoContext } from '../_layout';

const SHADialog = ({ show, setShow }: IDialog<ISHADialogData>) => {
  const { account } = useParams();

  const { repoName } = useOutletContext<IRepoContext>();

  return null;

  const url = `${registryHost}/${account}/${repoName}:${
    show?.data?.tag ? show?.data?.tag : `@${show?.data?.sha}`
  }`;
  return (
    <Popup.Root show={show as any} onOpenChange={setShow}>
      <Popup.Header>
        <div className="flex flex-row items-center gap-2xl">
          <Tag size={20} />
          {!!show?.data?.tag && <span>{show?.data?.tag}</span>}
          {!show?.data?.tag && <span>SHA256</span>}
        </div>
      </Popup.Header>
      <Popup.Content>
        <div className="flex flex-col gap-lg">
          {show?.data?.tag && (
            <code className="break-all">{show?.data?.sha}</code>
          )}
          <div className="flex flex-col gap-4xl rounded border border-border-default p-3xl">
            <div className="flex flex-col gap-lg">
              <div className="text-text-soft flex flex-row items-center gap-lg">
                <TerminalWindow size={20} weight={3} />
                <span>Install from the command line</span>
              </div>
              <CodeView showShellPrompt copy data={`docker pull ${url}`} />
            </div>
            <div className="flex flex-col gap-lg">
              <div className="text-text-soft flex flex-row items-center gap-lg">
                <CodeSimpleFill size={20} />
                <span>Use as base image in Dockerfile:</span>
              </div>
              <CodeView copy data={`FROM ${url}`} />
            </div>
          </div>
        </div>
      </Popup.Content>
    </Popup.Root>
  );
};

export default SHADialog;
