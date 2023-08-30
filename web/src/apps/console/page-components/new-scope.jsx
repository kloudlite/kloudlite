import { useParams } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IdSelector, idTypes } from '~/console/components/id-selector';
import { useReload } from '~/root/lib/client/helpers/reloader';
import {
  getMetadata,
  parseDisplaynameFromAnn,
  parseName,
  parseTargetNamespce,
} from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import * as Chips from '~/components/atoms/chips';
import { toast } from '~/components/molecule/toast';
import { useEffect, useState } from 'react';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import {
  getWorkspace,
  getWorkspaceSpecs,
} from '~/console/server/r-urils/workspace';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';

export const SCOPE = Object.freeze({
  ENVIRONMENT: 'environment',
  WORKSPACE: 'workspace',
});

const HandleScope = ({ show, setShow, scope }) => {
  const api = useAPIClient();
  const reloadPage = useReload();

  const { project: projectName } = useParams();
  const project = useDataFromMatches('project', {});
  const user = useDataFromMatches('user', {});

  const [validationSchema, setValidationSchema] = useState(
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
            env: getWorkspace({
              metadata: getMetadata({
                name: val.name,
                namespace: parseTargetNamespce(project),
                annotations: {
                  [keyconstants.author]: user.name,
                },
              }),
              displayName: val.displayName,
              spec: getWorkspaceSpecs({
                projectName,
                targetNamespace: `${projectName}-${val.name}`,
              }),
            }),
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
            secret: getWorkspace({
              metadata: getMetadata({
                namespace: projectName,
                name: parseName(show.data),
                annotations: {
                  [keyconstants.displayName]: val.displayName,
                  [keyconstants.author]: user.name,
                },
              }),
              spec: getWorkspaceSpecs({
                targetNamespace: projectName,
                projectName,
              }),
            }),
          });
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        setShow(false);
        resetValues();
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  useEffect(() => {
    if (show?.type === 'edit') {
      setValues((v) => ({
        ...v,
        displayName: parseDisplaynameFromAnn(show.data),
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
      show={show}
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
                  type: Chips.ChipType.BASIC,
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
                  scope === SCOPE.ENVIRONMENT
                    ? idTypes.environment
                    : idTypes.workspace
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
