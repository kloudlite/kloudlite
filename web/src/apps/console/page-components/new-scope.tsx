import { useParams } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IdSelector } from '~/console/components/id-selector';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { parseName, parseTargetNs } from '~/console/server/r-utils/common';
import * as Chips from '~/components/atoms/chips';
import { toast } from '~/components/molecule/toast';
import { useEffect, useState } from 'react';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import { handleError } from '~/root/lib/utils/common';
import { useConsoleApi } from '../server/gql/api-provider';
import { IHandleProps } from '../server/utils/common';
import { IWorkspace } from '../server/gql/queries/workspace-queries';

export const SCOPE = Object.freeze({
  ENVIRONMENT: 'environment',
  WORKSPACE: 'workspace',
});

const HandleScope = ({
  show,
  setShow,
  scope,
}: IHandleProps<{
  type: 'add' | 'edit';
  data: null | IWorkspace;
} | null> & {
  scope: 'environment' | 'workspace';
}) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { project: projectName } = useParams();
  const project = useDataFromMatches('project', {});

  const [validationSchema, setValidationSchema] = useState<any>(
    Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
    })
  );

  const {
    values,
    errors,
    handleSubmit,
    handleChange,
    isLoading,
    resetValues,
    setValues,
  } = useForm({
    initialValues: {
      name: '',
      displayName: '',
    },
    validationSchema,

    onSubmit: async (val) => {
      try {
        if (show?.type === 'add') {
          const createApi =
            scope === SCOPE.ENVIRONMENT
              ? api.createEnvironment
              : api.createWorkspace;
          const { errors: e } = await createApi({
            env: {
              metadata: {
                name: val.name,
                namespace: parseTargetNs(project),
              },
              displayName: val.displayName,
              spec: {
                projectName: projectName || '',
                targetNamespace: `${projectName}-${val.name}`,
              },
            },
          });
          if (e) {
            throw e[0];
          }
          toast.success('workspace created successfully');
        } else {
          const updateApi =
            scope === SCOPE.ENVIRONMENT
              ? api.updateEnvironment
              : api.updateWorkspace;
          const { errors: e } = await updateApi({
            env: {
              metadata: {
                namespace: projectName,
                name: parseName(show?.data),
              },
              displayName: val.displayName,
              spec: {
                targetNamespace: `${projectName}=${val.name}`,
                projectName: projectName || '',
              },
            },
          });
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        setShow(null);
        resetValues();
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (show && show.type === 'edit') {
      setValues((v) => ({
        ...v,
        displayName: show.data?.displayName || '',
      }));
      setValidationSchema(
        Yup.object({
          displayName: Yup.string().trim().required(),
        })
      );
    }
  }, [show]);

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
        {show?.type === 'add' ? `Create new ${scope}` : `Edit ${scope}`}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            {show?.type === 'edit' && (
              <Chips.Chip
                {...{
                  item: { id: parseName(show.data) },
                  label: parseName(show.data),
                  prefix: 'Id:',
                  disabled: true,
                  type: 'BASIC',
                }}
              />
            )}

            <TextInput
              label="Name"
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
              value={values.displayName}
              name="provider-secret-name"
            />
            {show?.type === 'add' && (
              <IdSelector
                name={values.displayName}
                resType={
                  scope === SCOPE.ENVIRONMENT ? 'environment' : 'workspace'
                }
                onChange={(id) => {
                  handleChange('name')({ target: { value: id } });
                }}
              />
            )}
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content={show?.type === 'add' ? 'Add' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleScope;
