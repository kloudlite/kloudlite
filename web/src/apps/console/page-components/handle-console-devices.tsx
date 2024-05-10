/* eslint-disable no-nested-ternary */
/* eslint-disable react/destructuring-assignment */
import {
  ArrowLineDown,
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  Plus,
  SmileySad,
  X,
} from '~/console/components/icons';
import { useEffect, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import { NumberInput } from '~/components/atoms/input';
import { usePagination } from '~/components/molecule/pagination';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { cn, useMapper } from '~/components/utils';
import List from '~/console/components/list';
import NoResultsFound from '~/console/components/no-results-found';
import QRCode from '~/console/components/qr-code';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/console/components/types.d';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { downloadFile } from '~/console/utils/commons';
import CodeView from '~/console/components/code-view';
import { InfoLabel } from '~/console/components/commons';
import { parseValue } from '~/console/page-components/util';
import { NameIdView } from '~/console/components/name-id-view';
import { IConsoleDevice } from '~/console/server/gql/queries/console-vpn-queries';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import Select from '~/components/atoms/select';
import { ConsoleApiType } from '../server/gql/saved-queries';
import ExtendedFilledTab from '../components/extended-filled-tab';
import { LoadingIndicator, LoadingPlaceHolder } from '../components/loading';

interface IExposedPorts {
  targetPort?: number;
  port?: number;
}

interface IExposedPortList {
  exposedPorts: IExposedPorts[];
  onDelete: (exposedPorts: IExposedPorts) => void;
}
const ExposedPortList = ({
  exposedPorts,
  onDelete = (_) => _,
}: IExposedPortList) => {
  const itemsPerPage = 4;

  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: exposedPorts,
      itemsPerPage,
    });

  useEffect(() => {
    setItems(exposedPorts);
  }, [exposedPorts]);
  return (
    <div className="flex flex-col gap-lg bg-surface-basic-default">
      {exposedPorts.length > 0 && (
        <List.Root
          className="min-h-[265px] !shadow-none"
          header={
            <div className="flex flex-row items-center w-full">
              <div className="text-text-strong bodyMd flex-1">
                Exposed ports
              </div>
              <div className="flex flex-row items-center">
                <IconButton
                  icon={<ChevronLeft />}
                  size="xs"
                  variant="plain"
                  onClick={() => onPrev()}
                  disabled={!hasPrevious}
                />
                <IconButton
                  icon={<ChevronRight />}
                  size="xs"
                  variant="plain"
                  onClick={() => onNext()}
                  disabled={!hasNext}
                />
              </div>
            </div>
          }
        >
          {page.map((ep, index) => {
            return (
              <List.Row
                className={cn({
                  '!border-b': index < itemsPerPage - 1,
                  '!rounded-b-none': index < itemsPerPage - 1,
                })}
                key={ep.port}
                columns={[
                  {
                    key: `${index}-column-2`,
                    className: 'flex-1',
                    render: () => (
                      <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                        <span>Exposed: </span>
                        {ep.port}
                        <ArrowRight size={16} />
                        <span>Target: </span>
                        {ep.targetPort}
                      </div>
                    ),
                  },
                  {
                    key: `${index}-column-3`,
                    render: () => (
                      <div>
                        <IconButton
                          icon={<X />}
                          variant="plain"
                          size="sm"
                          onClick={() => {
                            onDelete(ep);
                          }}
                        />
                      </div>
                    ),
                  },
                ]}
              />
            );
          })}
        </List.Root>
      )}
      {exposedPorts.length === 0 && (
        <div className="rounded border-border-default border min-h-[265px] flex flex-row items-center justify-center">
          <NoResultsFound
            title={null}
            subtitle="No ports are exposed currently"
            compact
            image={<SmileySad size={32} weight={1} />}
            shadow={false}
            border={false}
          />
        </div>
      )}
    </div>
  );
};

export const ExposedPorts = ({
  ports,
  onChange,
}: {
  ports: IExposedPorts[];
  onChange: (ports: IExposedPorts[]) => void;
}) => {
  const { errors, handleChange, submit, values, resetValues } = useForm({
    initialValues: {
      port: '',
      targetPort: '',
    },
    validationSchema: Yup.object({
      port: Yup.number()
        .required()
        .test('is-valid', 'Port already exists.', (value) => {
          return !ports.some((p) => p.port === value);
        }),
      targetPort: Yup.number().min(0).max(65535).required(),
    }),
    onSubmit: (val) => {
      onChange?.([
        ...ports,
        {
          port:
            typeof val.port === 'string' ? parseInt(val.port, 10) : val.port,
          targetPort:
            typeof val.targetPort === 'string'
              ? parseInt(val.targetPort, 10)
              : val.targetPort,
        },
      ]);
      resetValues();
    },
  });

  return (
    <>
      <div className="flex flex-col gap-3xl">
        <div className="flex flex-row gap-3xl items-start">
          <div className="flex-1">
            <NumberInput
              label={
                <InfoLabel label="Expose Port" info="info about expose port" />
              }
              size="lg"
              error={!!errors.port}
              message={errors.port}
              value={values.port}
              onChange={({ target }) => {
                handleChange('port')(dummyEvent(parseValue(target.value, 0)));
              }}
            />
          </div>
          <div className="flex-1">
            <NumberInput
              min={0}
              max={65536}
              label={
                <InfoLabel info="info about target port" label="Target port" />
              }
              size="lg"
              autoComplete="off"
              value={values.targetPort}
              onChange={({ target }) => {
                handleChange('targetPort')(
                  dummyEvent(parseValue(target.value, 0))
                );
              }}
            />
          </div>
          <div className="flex pt-5xl">
            <IconButton
              icon={<Plus />}
              variant="basic"
              disabled={!values.port || !values.targetPort}
              onClick={submit}
            />
          </div>
        </div>
      </div>
      <ExposedPortList
        exposedPorts={ports}
        onDelete={(ep) => {
          onChange?.(ports.filter((v) => v.port !== ep.port));
        }}
      />
    </>
  );
};

export const QRCodeView = ({ data }: { data: string }) => {
  return (
    <div className="flex flex-row gap-7xl">
      <div className="flex flex-col gap-2xl">
        <div className="bodyLg-medium text-text-default">
          Use WireGuard on your phone
        </div>
        <ul className="flex flex-col gap-lg bodyMd text-text-default list-disc list-outside pl-2xl">
          <li>Download the app from Google Play or Apple Store</li>
          <li>Open the app on your Phone</li>
          <li>Tab on the ➕ Plus icon</li>
          <li>Point your phone to this screen to capture the QR code</li>
        </ul>
      </div>
      <div>
        <QRCode value={data} />
      </div>
    </div>
  );
};

export const decodeConfig = ({
  encoding,
  value,
}: {
  encoding: string;
  value: string;
}) => {
  switch (encoding) {
    case 'base64':
      return atob(value);
    default:
      return value;
  }
};

const downloadConfig = ({
  filename,
  data,
}: {
  filename: string;
  data: string;
}) => {
  downloadFile({ filename, data, format: 'text/plain' });
};

export const ShowWireguardConfig = ({
  visible,
  setVisible,
  deviceName,
}: // data,
{
  visible: boolean;
  setVisible: (visible: boolean) => void;
  deviceName: string;
}) => {
  const [mode, setMode] = useState<'config' | 'qr'>('qr');

  const [data, setData] = useState<{
    value: string;
    encoding: string;
  }>();

  const api = useConsoleApi();

  const { data: devData, isLoading } = useCustomSwr(
    `device-${deviceName}`,
    async () => api.getGlobalVpnDevice({ deviceName, gvpn: 'default' }),
    true
  );

  useEffect(() => {
    setData(devData?.wireguardConfig);
  }, [devData]);

  const modeView = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col items-center justify-center">
          <LoadingPlaceHolder />
        </div>
      );
    }

    if (!data) {
      return (
        <div className="h-[100px] flex items-center justify-center">
          No wireguard config found.
        </div>
      );
    }

    const config = decodeConfig(data);
    switch (mode) {
      case 'qr':
        return <QRCodeView data={config} />;
      case 'config':
      default:
        return (
          <div className="flex flex-col gap-3xl">
            <div className="bodyMd text-text-default">
              Please use the following configuration to set up your WireGuard
              client.
            </div>
            <CodeView data={config} showShellPrompt={false} copy />
          </div>
        );
    }
  };

  return (
    <Popup.Root show={visible} onOpenChange={setVisible}>
      <Popup.Header>
        {mode === 'config' ? 'Wireguard Config' : 'Wireguard Config QR Code'}
      </Popup.Header>
      <Popup.Content>
        <div>
          <ExtendedFilledTab
            value={mode}
            onChange={(v) => {
              setMode(v as any);
            }}
            items={[
              {
                label: 'QR Code',
                value: 'qr',
              },
              {
                label: 'Config',
                value: 'config',
              },
            ]}
          />

          {modeView()}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button
          onClick={() => {
            if (!data) {
              toast.error('No wireguard config found.');
              return;
            }

            downloadConfig({
              filename: `wireguardconfig.yaml`,
              data: decodeConfig(data),
            });
          }}
          content="Export"
          prefix={<ArrowLineDown />}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

export const switchEnvironment = async ({
  api,
  device,
  environment,
}: {
  api: ConsoleApiType;
  device: IConsoleDevice;
  environment: string;
}) => {
  try {
    const { errors } = await api.updateConsoleVpnDevice({
      vpnDevice: {
        displayName: device.displayName,
        metadata: {
          name: parseName(device),
        },
        environmentName: environment,

        spec: {
          ports: device.spec?.ports,
        },
      },
    });
    if (errors) {
      throw errors[0];
    }
    toast.success('Device switched successfully');
  } catch (err) {
    handleError(err);
  }
};

type IDialog = IDialogBase<IConsoleDevice>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: parseName(props.data),
            ports: props.data.spec?.ports || [],
            isNameError: false,
            environmentName: props.data.environmentName,
          }
        : {
            displayName: '',
            name: '',
            ports: [],
            isNameError: false,
            projectName: '',
            environmentName: '',
          },
      validationSchema: isUpdate
        ? Yup.object({
            name: Yup.string().required(),
            displayName: Yup.string().required(),
            projectName: Yup.string().required(),
            environmentName: Yup.string().required(),
          })
        : Yup.object({
            name: Yup.string().required(),
            displayName: Yup.string().required(),
          }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            const { errors } = await api.createConsoleVpnDevice({
              vpnDevice: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                spec: {
                  ports: val.ports,
                },
              },
            });
            if (errors) {
              throw errors[0];
            }
            toast.success('Device created successfully');
          } else if (isUpdate && props.data) {
            const { errors } = await api.updateConsoleVpnDevice({
              vpnDevice: {
                displayName: val.displayName,
                metadata: {
                  name: parseName(props.data),
                },
                environmentName: val.environmentName,
                spec: {
                  ports: val.ports,
                },
              },
            });
            if (errors) {
              throw errors[0];
            }
            toast.success('Device updated successfully');
          }
          reloadPage();
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

  useEffect(() => {
    if (!isUpdate) {
      resetValues();
    }
  }, []);

  const { data: envData, isLoading: envLoading } = useCustomSwr(
    () => (values.projectName ? `/environments-${values.projectName}` : null),
    async () => {
      if (!values.projectName) {
        throw new Error('Project name is required!.');
      }
      return api.listEnvironments({});
    }
  );

  const environments = useMapper(parseNodes(envData), (val) => ({
    label: val.displayName,
    value: parseName(val),
    project: val,
    render: () => (
      <div className="flex flex-col">
        <div>{val.displayName}</div>
        <div className="bodySm text-text-soft">{parseName(val)}</div>
      </div>
    ),
  }));

  useEffect(() => {
    console.log('errors here', errors);
  }, [errors]);

  return (
    <Popup.Form
      onSubmit={(e) => {
        console.log('name error', values.isNameError);

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
            resType="console_vpn_device"
            displayName={values.displayName}
            name={values.name}
            label="Device name"
            placeholder="Enter device name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
          />
          {isUpdate && (
            <>
              <div className="flex flex-row items-start gap-3xl">
                <div className="basis-full">
                  <Select
                    label="Environment"
                    size="lg"
                    placeholder="Select a environment"
                    error={!!values.projectName && !!errors.environmentName}
                    message={values.projectName ? errors.environmentName : ''}
                    disabled={!values.projectName}
                    disableWhileLoading
                    loading={envLoading}
                    options={async () => [...environments]}
                    value={values.environmentName}
                    onChange={(val) => {
                      handleChange('environmentName')(dummyEvent(val.value));
                    }}
                  />
                </div>
              </div>
              <ExposedPorts
                ports={values.ports}
                onChange={(ports) => {
                  handleChange('ports')(dummyEvent(ports));
                }}
              />
            </>
          )}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          type="submit"
          content={!isUpdate ? 'Create' : 'Update'}
          variant="primary"
          loading={isLoading}
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleConsoleDevices = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Add device"
      updateTitle="Device setup"
      root={Root}
    />
  );
};

export default HandleConsoleDevices;
