/* eslint-disable react/destructuring-assignment */
import { useCallback, useEffect, useState } from 'react';
import Radio from '@kloudlite/design-system/atoms/radio';
import Select from '@kloudlite/design-system/atoms/select';
import Popup from '@kloudlite/design-system/molecule/popup';
import { toast } from '@kloudlite/design-system/molecule/toast';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { NameIdView } from '~/console/components/name-id-view';
import { IDialogBase } from '~/console/components/types.d';
import { findClusterStatus } from '~/console/hooks/use-cluster-status';
import { ClusterSelectItem } from '~/console/page-components/handle-environment';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IEnvironments } from '~/console/server/gql/queries/environment-queries';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

type IDialog = IDialogBase<ExtractNodeType<IEnvironments>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const [clusterList, setClusterList] = useState<any[]>([]);

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
  }, []);

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    resetValues,
    setValues,
    isLoading,
  } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      radioType: 'compute',
      environmentRoutingMode: false,
      isNameError: false,
      clusterName: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required('Name is required.'),
      displayName: Yup.string().required(),
      clusterName: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      if (isUpdate) {
        try {
          const { errors: e } = await api.cloneEnvironment({
            displayName: val.displayName,
            environmentRoutingMode: val.environmentRoutingMode
              ? 'public'
              : 'private',
            destinationEnvName: val.name,
            clusterName:
              val.radioType === 'template' ? '' : val.clusterName || '',
            sourceEnvName: parseName(props.data),
          });
          if (e) {
            throw e[0];
          }
          resetValues();
          toast.success('Environment cloned successfully');
          setVisible(false);
          reloadPage();
        } catch (err) {
          handleError(err);
        }
      }
    },
  });

  useEffect(() => {
    if (clusterList.length > 0) {
      setValues((v) => ({
        ...v,
        clusterName: clusterList.find((c) => c.ready)?.value || '',
      }));
    }
  }, [clusterList]);

  return (
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
            displayName={values.displayName}
            name={values.name}
            resType="environment"
            errors={errors.name}
            label="Name"
            placeholder="Environment name"
            handleChange={handleChange}
            nameErrorLabel="isNameError"
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
              label="Select Cluster"
              size="lg"
              value={values.clusterName}
              placeholder="Select a Cluster"
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
          {/*       Public environments will expose services to the public internet. */}
          {/*       Private environments will be accessible when Kloudlite VPN is */}
          {/*       active. */}
          {/*     </span> */}
          {/*   } */}
          {/* /> */}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          type="submit"
          content="Clone"
          variant="primary"
          loading={isLoading}
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const CloneEnvironment = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      root={Root}
      updateTitle="Clone Environment"
      createTitle="Clone Environment"
    />
  );
};

export default CloneEnvironment;
