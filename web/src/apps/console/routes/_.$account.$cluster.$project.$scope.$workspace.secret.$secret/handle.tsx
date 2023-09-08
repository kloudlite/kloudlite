import { TextArea, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import {
  parseFromAnn,
  parseName,
  parseTargetNs,
} from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { SecretIn } from '~/root/src/generated/gql/server';
import { IDialog } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/app-dialogs';

type IDialogValue = {
  key: string;
  value: string;
};

interface IUpdateSecret {
  api: ConsoleApiType;
  context: any;
  secret: SecretIn;
  data: any;
  reload: () => void;
}

export const updateSecret = async ({
  api,
  context,
  secret,
  data,
  reload,
}: IUpdateSecret) => {
  const { workspace, user } = context;

  try {
    const { errors: e } = await api.updateSecret({
      secret: {
        metadata: {
          name: parseName(secret),
          namespace: parseTargetNs(workspace),
          annotations: {
            [keyconstants.author]: user?.name || '',
            [keyconstants.node_type]: parseFromAnn(
              secret,
              keyconstants.node_type
            ),
          },
        },
        displayName: secret.displayName,
        stringData: data,
      },
    });
    if (e) {
      throw e[0];
    }
    reload();
  } catch (err) {
    handleError(err);
  }
};

export const ManageSecretDialog = ({
  show,
  setShow,
  onSubmit,
}: IDialog<IDialogValue>) => {
  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        key: '',
        value: '',
      },
      validationSchema: Yup.object({
        key: Yup.string()
          .required()
          .test('is-valid', 'Key already exists.', (value) => {
            return !show?.data?.[value];
          }),
        value: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        try {
          if (onSubmit) {
            onSubmit(val);
          }
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Root
      show={show as any}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>Add new entry</Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Key"
              value={values.key}
              onChange={handleChange('key')}
              error={!!errors.key}
              message={errors.key}
            />
            <TextArea
              label="Value"
              value={values.value}
              onChange={handleChange('value')}
              error={!!errors.value}
              message={errors.value}
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content="Create"
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};
