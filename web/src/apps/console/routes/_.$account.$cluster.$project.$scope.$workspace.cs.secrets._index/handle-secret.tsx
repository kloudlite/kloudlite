import { useOutletContext } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { IdSelector } from '~/console/components/id-selector';
import { parseTargetNs } from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import { IHandleProps } from '~/console/server/utils/common';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';

interface IData {
  type: 'add';
}

const Main = ({ show, setShow }: IHandleProps<IData>) => {
  const api = useAPIClient();
  const { workspace, user } = useOutletContext<IWorkspaceContext>();
  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        displayName: '',
        name: '',
        nodeType: '',
      },
      validationSchema: Yup.object({
        displayName: Yup.string().required(),
        name: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.createSecret({
            secret: {
              metadata: {
                name: val.name,
                namespace: parseTargetNs(workspace),
                annotations: {
                  [keyconstants.author]: user.name,
                  [keyconstants.nodeType]: val.nodeType,
                },
              },
              displayName: val.displayName,
              data: {},
            },
          });
          if (e) {
            throw e[0];
          }
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Root
      show={!!show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>
        {show?.type === 'add' ? 'Add new secret' : 'Edit secret'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              value={values.displayName}
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
            />
            <IdSelector
              resType="secret"
              onChange={(v) => handleChange('name')(dummyEvent(v))}
              name={values.displayName}
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content={show?.type === 'add' ? 'Create' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

const HandleSecret = ({ show, setShow }: IHandleProps<IData>) => {
  if (show) {
    return <Main show={show} setShow={setShow} />;
  }
  return null;
};

export default HandleSecret;
