import { Buildings, CopySimple } from '~/console/components/icons';
import { useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
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
import { parseName } from '~/console/server/r-utils/common';

import SecondarySubHeader from '~/console/components/secondary-sub-header';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import { IAccount } from '~/console/server/gql/queries/account-queries';
import DeleteDialog from '~/console/components/delete-dialog';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { IAccountContext } from '../_layout';
// import SubNavAction from '../components/sub-nav-action';
// import { useConsoleApi } from '../server/gql/api-provider';
// import { IAccount } from '../server/gql/queries/access-queries';
// import { ConsoleApiType } from '../server/gql/saved-queries';
// import { IAccountContext } from './_.$account';
// import SecondarySubHeader from '../components/secondary-sub-header';

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
          name: parseName(data),
        },
        contactEmail: data.contactEmail,
        kloudliteGatewayRegion: data.kloudliteGatewayRegion,
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
  const [deleteAccount, setDeleteAccount] = useState(false);

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const api = useConsoleApi();
  const reload = useReload();
  const navigate = useNavigate();

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

  useEffect(() => {
    resetValues();
  }, [account]);

  return (
    <div className="flex flex-col gap-6xl">
      <SecondarySubHeader
        title="General"
        action={
          values.displayName !== account.displayName && (
            <div className="flex flex-row gap-3xl items-center">
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
                  if (!isLoading) submit();
                }}
                loading={isLoading}
              />
            </div>
          )
        }
      />

      <div className="flex flex-col gap-6xl">
        <Box title="Profile">
          <div className="flex flex-row items-center gap-3xl">
            <Avatar size="lg" color="one" image={<Buildings />} />{' '}
            <Button content="Upload photo" variant="basic" />
          </div>
          <div className="flex flex-row gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Account name"
                value={values.displayName}
                onChange={handleChange('displayName')}
              />
            </div>
            <div className="flex-1">
              <TextInput
                value={account.kloudliteGatewayRegion}
                label="Kloudlite gateway region"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      onClick={() => copy(account.kloudliteGatewayRegion)}
                      className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                      tabIndex={-1}
                      aria-label="copy account id"
                    >
                      <CopySimple size={16} />
                    </button>
                  </div>
                }
                disabled
              />
            </div>
          </div>
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Team URL"
                value={`${consoleBaseUrl}/${parseName(account)}`}
                message="This is your URL namespace within Kloudlite"
                disabled
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      onClick={() =>
                        copy(`consoleBaseUrl}/${parseName(account)}`)
                      }
                      className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                      tabIndex={-1}
                      aria-label="copy account url"
                    >
                      <CopySimple size={16} />
                    </button>
                  </div>
                }
              />
            </div>
            <div className="flex-1">
              <TextInput
                value={parseName(account)}
                label="Account ID"
                message="Used when interacting with the Kloudlite API"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      onClick={() => copy(parseName(account))}
                      className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                      tabIndex={-1}
                      aria-label="copy account id"
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

        <DeleteContainer
          title="Delete Account"
          action={async () => {
            setDeleteAccount(true);
          }}
        >
          Permanently remove your Account and all of its contents from the
          Kloudlite platform. This action is not reversible — please continue
          with caution.
        </DeleteContainer>
        <DeleteDialog
          resourceName={parseName(account)}
          resourceType="account"
          show={deleteAccount}
          setShow={setDeleteAccount}
          onSubmit={async () => {
            try {
              const { errors } = await api.deleteAccount({
                accountName: parseName(account),
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`Account deleted successfully`);
              setDeleteAccount(false);
              navigate(`/`);
            } catch (err) {
              handleError(err);
            }
          }}
        />
      </div>
    </div>
  );
};
export default SettingGeneral;
