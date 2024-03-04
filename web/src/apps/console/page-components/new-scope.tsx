import { useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Switch } from '~/components/atoms/switch';
import { Checkbox } from '~/components/atoms/checkbox';
import { InfoLabel } from '~/console/components/commons';
import { IDialog } from '../components/types.d';
import { useConsoleApi } from '../server/gql/api-provider';
import { DIALOG_TYPE } from '../utils/commons';
import { IEnvironment } from '../server/gql/queries/environment-queries';
import { NameIdView } from '../components/name-id-view';

const HandleScope = ({ show, setShow }: IDialog<IEnvironment | null>) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { project: projectName } = useParams();

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
      environmentRoutingMode: false,
      isNameError: false,
    },
    validationSchema,

    onSubmit: async (val) => {
      if (!projectName) {
        throw new Error('Project name is required!.');
      }
      try {
        if (show?.type === DIALOG_TYPE.ADD) {
          const { errors: e } = await api.createEnvironment({
            projectName,
            env: {
              metadata: {
                name: val.name,
              },
              displayName: val.displayName,
              spec: {
                projectName: projectName || '',
                routing: {
                  mode: val.environmentRoutingMode ? 'public' : 'private',
                },
              },
            },
          });
          if (e) {
            throw e[0];
          }
          toast.success('Environment created successfully');
        } else {
          const { errors: e } = await api.updateEnvironment({
            projectName,
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
    if (show && show.type === DIALOG_TYPE.EDIT) {
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
        {show?.type === DIALOG_TYPE.ADD
          ? `Create new environment`
          : `Edit environment`}
      </Popup.Header>
      <Popup.Form
        onSubmit={(e) => {
          if (!values.isNameError) {
            handleSubmit(e);
          } else {
            e.preventDefault();
          }
        }}
      >
        <Popup.Content>
          <div className="flex flex-col gap-3xl">
            <NameIdView
              resType="environment"
              label="Name"
              displayName={values.displayName}
              name={values.name}
              errors={errors.values}
              handleChange={handleChange}
              nameErrorLabel="isNameError"
              isUpdate={show?.type !== DIALOG_TYPE.ADD}
            />
            <div className="flex flex-row items-center gap-lg">
              <Checkbox
                label="Public"
                checked={values.environmentRoutingMode}
                onChange={(val) => {
                  handleChange('environmentRoutingMode')(dummyEvent(val));
                }}
              />
              <InfoLabel
                info={
                  <div>
                    <div className="bodyMd-medium">Public:</div>
                    <p>
                      Public environments will expose services to the public
                      internet.
                    </p>
                    <div className="bodyMd-medium">Private:</div>
                    <p>
                      Private environments will be accessible when Kloudlite VPN
                      is active.
                    </p>
                  </div>
                }
              />
            </div>
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content={show?.type === DIALOG_TYPE.ADD ? 'Add' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </Popup.Form>
    </Popup.Root>
  );
};

export default HandleScope;
