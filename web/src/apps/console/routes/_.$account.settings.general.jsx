import { CopySimple } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { Avatar } from '~/components/atoms/avatar';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { useLog } from '~/root/lib/client/hooks/use-log';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';

const SettingGeneral = () => {
  const { account } = useOutletContext();
  useLog(account);
  return (
    <>
      <div className="rounded border border-border-default bg-surface-basic-default shadow-button p-3xl flex flex-col gap-3xl ">
        <div className="text-text-strong headingLg">Profile</div>
        <div className="flex flex-col gap-3xl">
          <div className="flex flex-row items-center gap-3xl">
            <Avatar size="lg" color="one" />{' '}
            <Button content="Upload photo" variant="basic" />
          </div>
          <TextInput
            value={account.displayName || account.name}
            label="Account name"
          />
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Team URL"
                value={`${consoleBaseUrl}/${account.name}`}
                message="This is your URL namespace within Kloudlite"
              />
            </div>
            <div className="flex-1">
              <TextInput
                value={account.name}
                label="Account ID"
                message="Used when interacting with the Kloudlite API"
                suffixIcon={CopySimple}
              />
            </div>
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-critical bg-surface-basic-default shadow-button">
        <div className="text-text-strong headingLg">Delete Account</div>
        <div className="bodyMd text-text-default">
          Permanently remove your personal account and all of its contents from
          the Kloudlite platform. This action is not reversible, so please
          continue with caution.
        </div>
        <Button content="Delete" variant="critical" />
      </div>
    </>
  );
};
export default SettingGeneral;
