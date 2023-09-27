import { Buildings, CopySimple } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { Avatar } from '~/components/atoms/avatar';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { Box, DeleteContainer } from '../components/common-console-components';
import { IAccountContext } from './_.$account';

const SettingGeneral = () => {
  const { account } = useOutletContext<IAccountContext>();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  return (
    <>
      <Box title="Profile">
        <div className="flex flex-row items-center gap-3xl">
          <Avatar size="lg" color="one" image={<Buildings />} />{' '}
          <Button content="Upload photo" variant="basic" />
        </div>
        <TextInput value={account.displayName} label="Account name" />
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput
              label="Team URL"
              value={`${consoleBaseUrl}/${account.metadata.name}`}
              message="This is your URL namespace within Kloudlite"
              disabled
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    onClick={() =>
                      copy(`consoleBaseUrl}/${account.metadata.name}`)
                    }
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
            />
          </div>
          <div className="flex-1">
            <TextInput
              value={account.metadata.name}
              label="Account ID"
              message="Used when interacting with the Kloudlite API"
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    onClick={() => copy(account.metadata.name)}
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
              disabled
            />
          </div>
        </div>
      </Box>

      <DeleteContainer title="Delete Account" action={() => {}}>
        Permanently remove your personal account and all of its contents from
        the Kloudlite platform. This action is not reversible, so please
        continue with caution.
      </DeleteContainer>
    </>
  );
};
export default SettingGeneral;
