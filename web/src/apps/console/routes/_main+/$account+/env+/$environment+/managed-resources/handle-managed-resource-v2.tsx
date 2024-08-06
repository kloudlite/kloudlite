/* eslint-disable react/destructuring-assignment */
import Popup from '~/components/molecule/popup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/console/components/types.d';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import Select from '~/components/atoms/select';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useOutletContext, useParams } from '@remix-run/react';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useMapper } from '~/components/utils';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { LoadingPlaceHolder } from '~/console/components/loading';
import ListV2 from '~/console/components/listV2';
import { ListItem } from '~/console/components/console-list-components';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
import { useCallback, useEffect, useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import { NameIdView } from '~/console/components/name-id-view';
import { IImportedManagedResources } from '~/console/server/gql/queries/imported-managed-resource-queries';
import { IEnvironmentContext } from '../_layout';

type BaseType = ExtractNodeType<IImportedManagedResources>;
type IDialog = IDialogBase<ExtractNodeType<IImportedManagedResources>>;

const SelectItem = ({ label, value }: { label: string; value: string }) => {
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
  const { setVisible, isUpdate } = props;

  const api = useConsoleApi();
  const reloadPage = useReload();
  const { environment } = useOutletContext<IEnvironmentContext>();

  const [msvcList, setMsvcList] = useState<any[]>([]);

  const getMsvcs = useCallback(async () => {
    try {
      const msvcs = await api.listClusterMSvs({});
      const data = parseNodes(msvcs.data).map((c) => ({
        label: c.displayName,
        value: parseName(c),
        ready: true,
        render: () => <SelectItem label={c.displayName} value={parseName(c)} />,
      }));
      setMsvcList(data);
    } catch (err) {
      handleError(err);
    }
  }, []);

  useEffect(() => {
    getMsvcs();
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
    initialValues: isUpdate
      ? {
          isNameError: false,
        }
      : {
          isNameError: false,
          name: '',
          displayName: '',
          managedServiceName: '',
          managedResourceName: '',
        },
    validationSchema: Yup.object({
      managedServiceName: Yup.string().required(
        'integrated service is required'
      ),
      managedResourceName: Yup.string().required(
        'integrated resource name is required'
      ),
    }),
    onSubmit: async (val) => {
      try {
        if (!isUpdate) {
          const { errors: e } = await api.importManagedResource({
            envName: parseName(environment),
            msvcName: val.managedServiceName || '',
            mresName: val.managedResourceName || '',
            importName: val.name || '',
          });
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        resetValues();
        toast.success(
          `integrated resource ${
            isUpdate ? 'updated' : 'imported'
          } successfully`
        );
        setVisible(false);
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (msvcList.length > 0) {
      setValues((v) => ({
        ...v,
        managedServiceName: msvcList.find((c) => c.ready)?.value || '',
      }));
    }
  }, [msvcList]);

  const { data: mresData, isLoading: mresIsLoading } = useCustomSwr(
    () => `/managed-services${values.managedServiceName}`,
    async () => {
      return api.listManagedResources({
        search: {
          managedServiceName: {
            matchType: 'exact',
            exact: values.managedServiceName,
          },
        },
      });
    }
  );

  const mresList = useMapper(parseNodes(mresData), (item) => {
    return {
      label: item.displayName,
      value: parseName(item),
      ready: item.status?.isReady,
      render: () => (
        <SelectItem label={item.displayName} value={parseName(item)} />
      ),
    };
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
        <div className="flex flex-col gap-2xl">
          <NameIdView
            placeholder="Enter integrated service name"
            label="Name"
            resType="imported_managed_resource"
            name={values.name || ''}
            displayName={values.displayName || ''}
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
          />

          <Select
            label="Integrated Services"
            size="lg"
            value={values.managedServiceName}
            // disabled={msvcIsLoading}
            placeholder="Select a Integrated Service"
            options={async () => [
              ...((msvcList &&
                msvcList.filter((msvc) => {
                  return msvc.ready;
                })) ||
                []),
            ]}
            onChange={({ value }) => {
              handleChange('managedServiceName')(dummyEvent(value));
            }}
            error={!!errors.managedServiceName}
            message={errors.clusterName}
            // loading={msvcIsLoading}
          />

          <Select
            label="Integrated Resource"
            size="lg"
            value={values.managedResourceName}
            disabled={mresIsLoading}
            placeholder="Select a Integrated Resource"
            options={async () => [
              ...((mresList &&
                mresList.filter((mres) => {
                  return mres.ready;
                })) ||
                []),
            ]}
            onChange={({ value }) => {
              handleChange('managedResourceName')(dummyEvent(value));
            }}
            error={!!errors.managedResourceName}
            message={errors.managedResourceName}
            loading={mresIsLoading}
          />
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={isUpdate ? 'Update' : 'Import'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleManagedResourceV2 = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit External Name' : 'Import Integrated Resource'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleManagedResourceV2;

export const ViewSecret = ({
  show,
  setShow,
  item,
}: {
  show: boolean;
  setShow: () => void;
  item: BaseType;
}) => {
  const api = useConsoleApi();
  const { environment } = useOutletContext<IEnvironmentContext>();
  const [onYesClick, setOnYesClick] = useState(false);
  const params = useParams();
  ensureAccountClientSide(params);
  const { data, isLoading, error } = useCustomSwr(
    () => (onYesClick ? `secret _${item.secretRef?.name}` : null),
    async () => {
      if (!item.name) {
        toast.error('Secret not found');
        throw new Error('Secret not found');
      } else {
        return api.getSecret({
          envName: parseName(environment),

          name: item.secretRef?.name,
        });
      }
    }
  );

  const dataSecret = () => {
    if (isLoading) {
      return <LoadingPlaceHolder />;
    }
    if (error) {
      return (
        <span className="bodyMd-medium text-text-strong">
          Error while fetching secrets
        </span>
      );
    }
    if (!data?.stringData) {
      return (
        <span className="bodyMd-medium text-text-strong">No secret found</span>
      );
    }

    return (
      <ListV2.Root
        data={{
          headers: [
            {
              render: () => 'Key',
              name: 'key',
              className: 'min-w-[170px]',
            },
            {
              render: () => 'Value',
              name: 'value',
              className: 'flex-1 min-w-[345px] max-w-[345px] w-[345px]',
            },
          ],
          rows: Object.entries(data.stringData || {}).map(([key, value]) => {
            const v = value as string;
            return {
              columns: {
                key: {
                  render: () => <ListItem data={key} />,
                },
                value: {
                  render: () => (
                    <CopyContentToClipboard
                      content={v}
                      toastMessage={`${key} copied`}
                    />
                  ),
                },
              },
            };
          }),
        }}
      />
    );
  };

  useEffect(() => {
    if (error) {
      toast.error(error);
    }
  }, [error]);

  const { onNext, currentStep } = useMultiStep({
    defaultStep: 0,
    totalSteps: 2,
  });
  return (
    <Popup.Root show={show} onOpenChange={setShow}>
      <Popup.Header>
        {currentStep === 0 ? <div>Confirmation</div> : <div>Secrets</div>}
      </Popup.Header>
      <Popup.Content>
        <MultiStep.Root currentStep={currentStep}>
          <MultiStep.Step step={0}>
            <div>
              <p>{`Are you sure you want to view the secrets of '${item.name}'?`}</p>
            </div>
          </MultiStep.Step>
          <MultiStep.Step step={1}>{dataSecret()}</MultiStep.Step>
        </MultiStep.Root>
      </Popup.Content>
      <Popup.Footer>
        {currentStep === 0 ? (
          <Popup.Button
            content="Yes"
            onClick={() => {
              onNext();
              setOnYesClick(true);
            }}
          />
        ) : null}
      </Popup.Footer>
    </Popup.Root>
  );
};
