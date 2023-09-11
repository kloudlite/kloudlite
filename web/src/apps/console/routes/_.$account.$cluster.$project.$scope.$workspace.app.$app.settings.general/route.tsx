import { CopySimple } from '@jengaicons/react';
import { useEffect } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';

const SettingGeneral = () => {
  const { data: subNavData, setData: setSubNavAction } = useSubNavData();

  useEffect(() => {
    setSubNavAction({
      show: false,
      action: () => {
        console.log('done');
      },
    });
  }, []);

  return (
    <>
      <div className="rounded border border-border-default bg-surface-basic-default shadow-button p-3xl flex flex-col gap-3xl ">
        <div className="text-text-strong headingLg">Application Detail</div>
        <div className="flex flex-col gap-3xl">
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Application name"
                onChange={() => {
                  setSubNavAction({
                    ...(subNavData || {}),
                    show: true,
                    content: 'Commit n changes',
                  });
                }}
              />
            </div>
            <div className="flex-1">
              <TextInput
                label="Application ID"
                suffixIcon={<CopySimple />}
                onChange={() => {
                  setSubNavAction({
                    ...(subNavData || {}),
                    show: true,
                    content: 'Commit m changes',
                  });
                }}
              />
            </div>
          </div>
          <TextInput label="Description" />
        </div>
      </div>
      <div className="rounded border border-border-default bg-surface-basic-default shadow-button flex flex-col">
        <div className="flex flex-col gap-3xl p-3xl">
          <div className="text-text-strong headingLg">Transfer</div>
          <div className="bodyMd text-text-default">
            Move your app to a different workspace seamlessly, avoiding any
            downtime or disruptions to workflows.
          </div>
        </div>
        <div className="bg-surface-basic-subdued p-3xl flex flex-row justify-end">
          <Button variant="basic" content="Transfer" />
        </div>
      </div>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-critical bg-surface-basic-default shadow-button">
        <div className="text-text-strong headingLg">Delete Application</div>
        <div className="bodyMd text-text-default">
          Permanently remove your application and all of its contents from the
          “Lobster Early” project. This action is not reversible, so please
          continue with caution.
        </div>
        <Button content="Delete" variant="critical" />
      </div>
    </>
  );
};
export default SettingGeneral;
