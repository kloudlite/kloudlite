import { toast } from 'react-toastify';
import { TextArea, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import {
  getMetadata,
  parseDisplayname,
  parseFromAnn,
  parseName,
  parseTargetNamespce,
} from '~/console/server/r-urils/common';
import { getConfig } from '~/console/server/r-urils/config';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';

export const updateConfig = async ({ api, context, config, data, reload }) => {
  const { workspace, user } = context;
  try {
    const { errors: e } = await api.updateConfig({
      config: getConfig({
        metadata: getMetadata({
          name: parseName(config),
          namespace: parseTargetNamespce(workspace),
          annotations: {
            [keyconstants.author]: user.name,
            [keyconstants.node_type]: parseFromAnn(
              config,
              keyconstants.node_type
            ),
          },
        }),
        displayName: parseDisplayname(config),
        data,
      }),
    });
    if (e) {
      throw e[0];
    }
    reload();
  } catch (err) {
    toast.error(err.message);
  }
};

const Main = ({ show, setShow, onSubmit }) => {
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
            return !show?.data[value];
          }),
        value: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        try {
          if (onSubmit) {
            onSubmit(val);
          }
        } catch (err) {
          toast.error(err.message);
        }
      },
    });

  return (
    <Popup.Root
      show={show}
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

const Handle = ({ show, setShow, onSubmit }) => {
  if (show) {
    return <Main show={show} setShow={setShow} onSubmit={onSubmit} />;
  }
  return null;
};

export default Handle;
