/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
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
import { useOutletContext } from '@remix-run/react';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useMapper } from '~/components/utils';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { LoadingPlaceHolder } from '~/console/components/loading';
import ListV2 from '~/console/components/listV2';
import { ListItem } from '~/console/components/console-list-components';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
import { useEffect } from 'react';
import { IManagedResources } from '~/console/server/gql/queries/managed-resources-queries';
import { IEnvironmentContext } from '../_layout';

type BaseType = ExtractNodeType<IManagedResources>;
type IDialog = IDialogBase<ExtractNodeType<IManagedResources>>;

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

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            // displayName: props.data.displayName,
            // name: parseName(props.data),
            // recordType: props.data.spec?.recordType,
            // record: props.data.spec?.record,
            isNameError: false,
          }
        : {
            // name: '',
            // displayName: '',
            // recordType: recordTypes[0].value,
            // record: '',
            isNameError: false,
            managedServiceName: '',
            managedResourceName: '',
          },
      validationSchema: Yup.object({
        managedServiceName: Yup.string().required(
          'managed service is required'
        ),
        managedResourceName: Yup.string().required(
          'managed resource name is required'
        ),
      }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            const { errors: e } = await api.importManagedResource({
              envName: parseName(environment),
              msvcName: val.managedServiceName || '',
              mresName: val.managedResourceName || '',
              //   externalApp: {
              //     displayName: val.displayName,
              //     metadata: {
              //       name: val.name,
              //     },
              //     spec: {
              //       recordType: validateExternalAppRecordType(
              //         val.recordType || ''
              //       ),
              //       record: val.record || '',
              //     },
              //   },
            });
            if (e) {
              throw e[0];
            }
          }
          reloadPage();
          resetValues();
          toast.success(
            `managed resource ${isUpdate ? 'updated' : 'imported'} successfully`
          );
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

  const { data: msvcData, isLoading: msvcIsLoading } = useCustomSwr(
    () => 'managed-services',
    async () => {
      return api.listClusterMSvs();
    }
  );

  const msvcList = useMapper(parseNodes(msvcData), (item) => {
    return {
      label: item.displayName,
      value: parseName(item),
      render: () => (
        <SelectItem label={item.displayName} value={parseName(item)} />
      ),
    };
  });

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
          <Select
            label="Select Managed Services"
            size="lg"
            value={values.managedServiceName}
            disabled={msvcIsLoading}
            placeholder="Select a Managed Service"
            options={async () => msvcList}
            onChange={({ value }) => {
              handleChange('managedServiceName')(dummyEvent(value));
            }}
            error={!!errors.managedServiceName}
            message={errors.clusterName}
            loading={msvcIsLoading}
          />

          <Select
            label="Select Managed Resource"
            size="lg"
            value={values.managedResourceName}
            disabled={mresIsLoading}
            placeholder="Select a Managed Resource"
            options={async () => mresList}
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
        {isUpdate ? 'Edit External Name' : 'Import Managed Resource'}
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
  const { data, isLoading, error } = useCustomSwr(
    () => `secret _${item.syncedOutputSecretRef?.metadata?.name}`,
    async () => {
      if (!item.syncedOutputSecretRef?.metadata?.name) {
        toast.error('Secret not found');
        throw new Error('Secret not found');
      } else {
        return api.getSecret({
          envName: parseName(environment),

          name: item.syncedOutputSecretRef?.metadata?.name,
        });
      }
    }
  );

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
            {data && (
              <div>
                <p>{`Are you sure you want to view the secrets of '${data.displayName}'?`}</p>
              </div>
            )}
          </MultiStep.Step>
          <MultiStep.Step step={1}>
            {isLoading ? (
              <LoadingPlaceHolder />
            ) : (
              data && (
                <ListV2.Root
                  data={{
                    headers: [
                      {
                        render: () => 'Key',
                        name: 'key',
                        className: 'min-w-[170px]',
                      },
                      {
                        render: () => '',
                        name: 'copy',
                        className: 'max-w-[120px]',
                      },
                      {
                        render: () => 'Value',
                        name: 'value',
                        className: 'flex-1',
                      },
                    ],
                    rows: Object.entries(data.stringData).map(
                      ([key, value]) => {
                        const v = value as string;
                        return {
                          columns: {
                            key: {
                              render: () => <ListItem data={key} />,
                            },
                            value: {
                              render: () => (
                                <ListItem data={v} className="w-[220px]" />
                              ),
                            },
                            copy: {
                              render: () => (
                                <CopyContentToClipboard
                                  content={v}
                                  toastMessage={`${key} copied`}
                                  label="Copy Secret"
                                />
                              ),
                            },
                          },
                        };
                      }
                    ),
                  }}
                />
              )
            )}
          </MultiStep.Step>
        </MultiStep.Root>
      </Popup.Content>
      <Popup.Footer>
        {currentStep === 0 ? (
          <Popup.Button content="Yes" onClick={() => onNext()} />
        ) : null}
      </Popup.Footer>
    </Popup.Root>
  );
};
