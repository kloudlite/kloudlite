import { useEffect } from 'react';
import { TextArea, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { IDialog, IModifiedItem } from '~/console/components/types.d';
import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import { parseName } from '~/console/server/r-utils/common';
import useForm from '~/lib/client/hooks/use-form';
import Yup from '~/lib/server/helpers/yup';
import { handleError } from '~/lib/utils/common';
import { SecretIn } from '~/root/src/generated/gql/server';

type IDialogValue = {
  key: string;
  value: string;
};

interface IUpdateSecret {
  api: ConsoleApiType;
  secret: SecretIn;
  environment: string;
  data: any;
  reload: () => void;
}

export const updateSecret = async ({
  api,
  secret,
  data,
  reload,
  environment,
}: IUpdateSecret) => {
  try {
    const { errors: e } = await api.updateSecret({
      envName: environment,

      secret: {
        metadata: {
          name: parseName(secret),
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

const Handle = ({
  show,
  setShow,
  onSubmit,
  isUpdate,
}: IDialog<IModifiedItem, IDialogValue> & { isUpdate?: boolean }) => {
  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    resetValues,
    isLoading,
    setValues,
  } = useForm({
    initialValues: {
      key: '',
      value: '',
    },
    validationSchema: Yup.object({
      key: Yup.string()
        .required()
        .test('is-valid', 'Key already exists.', (value) => {
          if (isUpdate) {
            return true;
          }
          return !show?.data?.[value];
        }),
      value: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      try {
        if (onSubmit) {
          const value = Object.values(show?.data || {})?.[0];
          onSubmit(val, value);
          resetValues();
        }
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (isUpdate && show) {
      const { data } = show;
      const key = Object.keys(data)?.[0];
      const value = Object.values(data)?.[0];
      const newVal = value.newvalue || value.value;
      // @ts-ignore
      setValues((v) => ({
        // @ts-ignore
        ...v,
        key,
        value: newVal,
      }));
    }
  }, [isUpdate, show]);

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
      <Popup.Header>{isUpdate ? 'Update entry' : 'Add new entry'}</Popup.Header>
      <Popup.Form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Key"
              value={values.key}
              onChange={handleChange('key')}
              error={!!errors.key}
              message={errors.key}
              disabled={isUpdate}
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
            content={isUpdate ? 'Update' : 'Create'}
            variant="primary"
          />
        </Popup.Footer>
      </Popup.Form>
    </Popup.Root>
  );
};

export default Handle;
