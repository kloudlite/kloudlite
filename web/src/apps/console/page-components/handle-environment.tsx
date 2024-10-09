/* eslint-disable no-nested-ternary */
import { useCallback, useEffect, useState } from 'react';
import Radio from '~/components/atoms/radio';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { cn } from '~/components/utils';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { NameIdView } from '../components/name-id-view';
import { IDialog } from '../components/types.d';
import { findClusterStatus } from '../hooks/use-cluster-status';
import { useConsoleApi } from '../server/gql/api-provider';
import { IEnvironment } from '../server/gql/queries/environment-queries';
import { parseName, parseNodes } from '../server/r-utils/common';
import { DIALOG_TYPE } from '../utils/commons';

export const ClusterSelectItem = ({
  label,
  value,
  disabled,
}: {
  label: string;
  value: string;
  disabled?: boolean;
}) => {
  return (
    <div className={cn({ 'cursor-default': !!disabled })}>
      <div className="flex flex-col">
        <div className={disabled ? 'text-text-disabled' : 'text-text-default'}>
          {label}
        </div>
        <div
          className={cn('bodySm text-text-default', {
            'text-text-disabled': !!disabled,
          })}
        >
          {value}
        </div>
      </div>
    </div>
  );
};

const HandleEnvironment = ({ show, setShow }: IDialog<IEnvironment | null>) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const [clusterList, setClusterList] = useState<any[]>([]);

  // const klCluster = {
  //   label: 'Kloudlite cluster',
  //   value: constants.kloudliteClusterName,
  //   ready: true,
  //   render: () => (
  //     <ClusterSelectItem label="Kloudlite cluster" value="kloudlite-cluster" />
  //   ),
  // };

  const getClusters = useCallback(async () => {
    try {
      const byokClusters = await api.listByokClusters({});
      const data = parseNodes(byokClusters.data).map((c) => ({
        label: c.displayName,
        value: parseName(c),
        ready: findClusterStatus(c),
        disabled: () => !findClusterStatus(c),
        render: ({ disabled }: { disabled: boolean }) => (
          <ClusterSelectItem
            label={c.displayName}
            value={parseName(c)}
            disabled={disabled}
          />
        ),
      }));
      setClusterList(data);
    } catch (err) {
      handleError(err);
    }
  }, []);

  useEffect(() => {
    getClusters();
  }, [show]);

  const [validationSchema] = useState<any>(
    Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      // clusterName: Yup.string().required(),
    }),
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
      radioType: 'compute',
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
              clusterName:
                val.radioType === 'template' ? '' : val.clusterName || '',
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
    setValues((v) => ({
      ...v,
      clusterName:
        clusterList.length > 0
          ? clusterList.find((c) => c.ready)?.value || ''
          : '',
    }));
  }, [clusterList, show]);

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
          ? values.radioType === 'compute'
            ? `Create new environment`
            : `Create new template`
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

            <Radio.Root
              direction="horizontal"
              value={values.radioType}
              onChange={(value) => {
                handleChange('radioType')(dummyEvent(value));
              }}
            >
              <Radio.Item value="compute">Environment</Radio.Item>
              <Radio.Item value="template">Environment Template</Radio.Item>
            </Radio.Root>

            {values.radioType === 'compute' && (
              <Select
                label="Select Compute"
                size="lg"
                value={values.clusterName}
                placeholder="Select a Compute"
                options={async () => clusterList}
                onChange={({ value }) => {
                  handleChange('clusterName')(dummyEvent(value));
                }}
                error={!!errors.clusterName}
                message={errors.clusterName}
              />
            )}

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

export default HandleEnvironment;
