import { TextArea, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { IDialog, IModifiedItem } from '~/console/components/types.d';
import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import { parseName } from '~/console/server/r-utils/common';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { ConfigIn } from '~/root/src/generated/gql/server';

export interface IConfigValue {
  key: string;
  value: string;
}

interface IUpdateConfig {
  api: ConsoleApiType;
  project: string;
  environment: string;
  config: ConfigIn;
  data: any;
  reload: () => void;
}

export const updateConfig = async ({
  api,
  project,
  environment,
  config,
  data,
  reload,
}: IUpdateConfig) => {
  try {
    const { errors: e } = await api.updateConfig({
      envName: environment,
      projectName: project,
      config: {
        metadata: {
          name: parseName(config),
        },
        displayName: config.displayName,
        data,
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

const Handle = ({
  show,
  setShow,
  onSubmit,
}: IDialog<IModifiedItem, IConfigValue>) => {
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
            resetValues();
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

export default Handle;
