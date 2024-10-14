import { useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import Popup from '~/components/molecule/popup';
import CodeView from '~/console/components/code-view';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';

export const LocalDeviceClusterInstructions = ({
  show,
  onClose,
}: {
  show: boolean;
  onClose: () => void;
}) => {
  const params = useParams();
  ensureAccountClientSide(params);

  return (
    <Popup.Root onOpenChange={onClose} show={show} className="!w-[800px]">
      <Popup.Header>Instructions to add your local device</Popup.Header>
      <Popup.Content>
        <form className="flex flex-col gap-2xl">
          <div className="flex flex-col gap-sm text-start ">
            <span className="flex flex-wrap items-center gap-md py-lg">
              1. Download and install kloudlite cli:
            </span>
            <CodeView
              preClassName="!overflow-none text-wrap break-words"
              copy
              data="curl 'https://kl.kloudlite.io/kloudlite/kl!?select=kl' | bash"
            />

            <span className="flex flex-wrap items-center gap-md py-lg">
              2. Login to your account:
            </span>
            <CodeView
              preClassName="!overflow-none text-wrap break-words"
              copy
              language="bash"
              data={`
kl auth login
# After login, you will be prompted to select your account.
                `.trim()}
            />

            <span className="flex flex-wrap items-center gap-md py-lg">
              3. Attach Compute:
            </span>
            <CodeView
              preClassName="!overflow-none text-wrap break-words"
              copy
              data="kl up"
            />

            {/* {data.url} */}
          </div>
        </form>
      </Popup.Content>
      <Popup.Footer>
        <Button variant="primary-outline" content="close" onClick={onClose} />
      </Popup.Footer>
    </Popup.Root>
  );
};
