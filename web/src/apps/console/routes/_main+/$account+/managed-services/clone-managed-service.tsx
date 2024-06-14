/* eslint-disable react/destructuring-assignment */
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { NameIdView } from '~/console/components/name-id-view';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IClusterMSvs } from '~/console/server/gql/queries/cluster-managed-services-queries';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import Select from '~/components/atoms/select';
import { useAppend, useMapper } from '~/components/utils';

type IDialog = IDialogBase<ExtractNodeType<IClusterMSvs>>;

const ClusterSelectItem = ({
  label,
  value,
}: {
  label: string;
  value: string;
}) => {
  return (
    <div>
      <div className="flex flex-col">
        <div>{label}</div>
        <div className="bodySm text-text-soft">{value}</div>
      </div>
    </div>
  );
};

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { data: clustersData, isLoading: cIsLoading } = useCustomSwr(
    'clusters',
    async () =>
      api.listClusters({
        pagination: {
          first: 100,
        },
      }),
    true
  );

  const { data: byokClustersData, isLoading: byokCIsLoading } = useCustomSwr(
    'byokclusters',
    async () =>
      api.listByokClusters({
        pagination: {
          first: 100,
        },
      }),
    true
  );

  const cData = useMapper(parseNodes(clustersData), (item) => {
    return {
      label: item.displayName,
      value: parseName(item),
      ready: item.status?.isReady,
      render: () => (
        <ClusterSelectItem label={item.displayName} value={parseName(item)} />
      ),
    };
  });

  const bCData = useMapper(parseNodes(byokClustersData), (item) => {
    return {
      label: item.displayName,
      value: parseName(item),
      ready: true,
      render: () => (
        <ClusterSelectItem label={item.displayName} value={parseName(item)} />
      ),
    };
  });

  const clusterList = useAppend(cData, bCData);

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        name: '',
        displayName: '',
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
            const { errors: e } = await api.cloneClusterMSv({
              displayName: val.displayName,
              destinationMsvcName: val.name,
              clusterName: val.clusterName,
              sourceMsvcName: parseName(props.data),
            });
            if (e) {
              throw e[0];
            }
            resetValues();
            toast.success('Managed service cloned successfully');
            setVisible(false);
            reloadPage();
          } catch (err) {
            handleError(err);
          }
        }
      },
    });
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
            placeholder="Managed service name"
            handleChange={handleChange}
            nameErrorLabel="isNameError"
          />

          <Select
            label="Select Cluster"
            size="lg"
            value={values.clusterName}
            disabled={cIsLoading}
            placeholder="Select a Cluster"
            options={async () => [
              ...((clusterList &&
                clusterList.filter((d) => {
                  return d.ready;
                })) ||
                []),
            ]}
            onChange={({ value }) => {
              handleChange('clusterName')(dummyEvent(value));
            }}
            error={!!errors.clusterName}
            message={errors.clusterName}
            loading={cIsLoading || byokCIsLoading}
          />
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

const CloneManagedService = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      root={Root}
      updateTitle="Clone ManagedService"
      createTitle="Clone ManagedService"
    />
  );
};

export default CloneManagedService;
