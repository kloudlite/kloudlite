import { Buildings, CopySimple } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { useEffect } from 'react';
import { Avatar } from '~/components/atoms/avatar';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Box, DeleteContainer } from '../components/common-console-components';
import GitRepoSelector from '../components/git-repo-selector';
import SubNavAction from '../components/sub-nav-action';
import { useConsoleApi } from '../server/gql/api-provider';
import { IAccount } from '../server/gql/queries/access-queries';
import { ConsoleApiType } from '../server/gql/saved-queries';
import { IAccountContext } from './_.$account';

export const updateAccount = async ({
  api,
  data,
}: {
  api: ConsoleApiType;
  data: IAccount;
}) => {
  try {
    const { errors: e } = await api.updateAccount({
      account: {
        displayName: data.displayName,
        metadata: {
          name: data.metadata.name,
        },
        contactEmail: data.contactEmail,
        spec: data.spec,
      },
    });
    if (e) {
      throw e[0];
    }
  } catch (err) {
    handleError(err);
  }
};

const SettingGeneral = () => {
  const { account } = useOutletContext<IAccountContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const api = useConsoleApi();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const { values, handleChange, submit, isLoading, resetValues } = useForm({
    initialValues: {
      displayName: account.displayName,
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required('Name is required.'),
    }),
    onSubmit: async (val) => {
      await updateAccount({
        api,
        data: { ...account, displayName: val.displayName },
      });
      resetAndReload();
    },
  });

  useEffect(() => {
    setHasChanges(values.displayName !== account.displayName);
  }, [values]);

  return (
    <>
      <SubNavAction deps={[values, isLoading]}>
        {values.displayName !== account.displayName && (
          <>
            <Button
              content="Discard"
              variant="basic"
              onClick={() => {
                resetValues();
              }}
            />
            <Button
              content="Save changes"
              variant="primary"
              onClick={() => {
                submit();
              }}
              loading={isLoading}
            />
          </>
        )}
      </SubNavAction>

      <GitRepoSelector />
      <Box title="Profile">
        <div className="flex flex-row items-center gap-3xl">
          <Avatar size="lg" color="one" image={<Buildings />} />{' '}
          <Button content="Upload photo" variant="basic" />
        </div>
        <TextInput
          label="Account name"
          value={values.displayName}
          onChange={handleChange('displayName')}
        />
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
