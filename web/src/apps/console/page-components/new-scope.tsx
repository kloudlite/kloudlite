import { useEffect, useState } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import Select from '~/components/atoms/select';
import { IDialog } from '../components/types.d';
import { useConsoleApi } from '../server/gql/api-provider';
import { DIALOG_TYPE } from '../utils/commons';
import { IEnvironment } from '../server/gql/queries/environment-queries';
import { NameIdView } from '../components/name-id-view';
import { parseName, parseNodes } from '../server/r-utils/common';

const HandleScope = ({ show, setShow }: IDialog<IEnvironment | null>) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { data: clustersData, isLoading: cIsLoading } = useCustomSwr(
    'clusters',
    async () => api.listClusters({}),
    true
  );

  const [validationSchema, setValidationSchema] = useState<any>(
    Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      clusterName: Yup.string().required(),
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
      clusterName: '',
      environmentRoutingMode: false,
      isNameError: false,
    },
    validationSchema,

    onSubmit: async (val) => {
      try {
        if (show?.type === DIALOG_TYPE.ADD) {
          const { errors: e } = await api.createEnvironment({
            env: {
              metadata: {
                name: val.name,
              },
              clusterName: val.clusterName || '',
              displayName: val.displayName,
              spec: {
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
            env: {
              metadata: {
                name: parseName(show?.data),
              },
              clusterName: show?.data?.clusterName || '',
              displayName: val.displayName,
              spec: {},
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

            <Select
              label="Select Cluster"
              size="lg"
              value={values.clusterName}
              disabled={cIsLoading}
              placeholder="Select a Cluster"
              options={async () => [
                ...((clustersData &&
                  parseNodes(clustersData)
                    .filter((d) => {
                      return d.status?.isReady;
                    })
                    .map((d) => ({
                      label: `${d.displayName} - ${parseName(d)}`,
                      value: parseName(d),
                    }))) ||
                  []),
              ]}
              onChange={({ value }) => {
                handleChange('clusterName')(dummyEvent(value));
              }}
              error={!!errors.clusterName}
              message={errors.clusterName}
            />

            {/* <Checkbox */}
            {/*   label="Public" */}
            {/*   checked={values.environmentRoutingMode} */}
            {/*   onChange={(val) => { */}
            {/*     handleChange('environmentRoutingMode')(dummyEvent(val)); */}
            {/*   }} */}
            {/* /> */}
            {/* <Banner */}
            {/*   type="info" */}
            {/*   body={ */}
            {/*     <span> */}
            {/*       Public environments will expose services to the public */}
            {/*       internet. Private environments will be accessible when */}
            {/*       Kloudlite VPN is active. */}
            {/*     </span> */}
            {/*   } */}
            {/* /> */}
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
