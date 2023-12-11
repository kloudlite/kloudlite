/* eslint-disable no-nested-ternary */
/* eslint-disable react/destructuring-assignment */
import {
  ArrowLineDown,
  ArrowRight,
  Check,
  ChevronLeft,
  ChevronRight,
  SmileySad,
  X,
} from '@jengaicons/react';
import { useOutletContext, useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import Chips from '~/components/atoms/chips';
import { NumberInput, TextInput } from '~/components/atoms/input';
import { usePagination } from '~/components/molecule/pagination';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { cn } from '~/components/utils';
import { IdSelector } from '~/console/components/id-selector';
import List from '~/console/components/list';
import NoResultsFound from '~/console/components/no-results-found';
import QRCode from '~/console/components/qr-code';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IDevices } from '~/console/server/gql/queries/vpn-queries';
import {
  ExtractNodeType,
  ensureResource,
  parseName,
} from '~/console/server/r-utils/common';
import { ensureClusterClientSide } from '~/console/server/utils/auth-utils';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import { ENV_NAMESPACE } from '~/root/lib/configs/env';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/console/components/types.d';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { LoadingPlaceHolder } from '~/console/components/loading';
import { downloadFile } from '~/console/utils/commons';
import CodeView from '~/console/components/code-view';
import { InfoLabel } from '~/console/components/commons';
import { parseValue } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/util';
import { IAccountContext } from '../_.$account';

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
                        <ArrowRight size={16} weight={1} />
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
  const { errors, handleChange, submit, values } = useForm({
    initialValues: {
      port: 3000,
      targetPort: 3000,
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
      onChange?.([...ports, val]);
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
              icon={<Check />}
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

const decodeConfig = ({
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
  data,
  mode = 'config',
}: {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  data: { device: string };
  mode: 'qr' | 'config';
}) => {
  const [loading, setLoading] = useState(false);
  const [config, setConfig] = useState<string | undefined>(undefined);
  const api = useConsoleApi();

  const { cluster } = useParams();

  useEffect(() => {
    if (visible) {
      (async () => {
        setLoading(true);
        try {
          const { errors, data: out } = await api.getVpnDevice({
            clusterName: cluster || '',
            name: data.device,
          });
          if (errors) {
            throw errors[0];
          }
          if (out.wireguardConfig) {
            setConfig(decodeConfig(out.wireguardConfig));
          } else {
            setConfig(undefined);
          }
        } catch (error) {
          handleError(error);
        } finally {
          setLoading(false);
        }
      })();
    }
  }, [visible]);

  const modeView = () => {
    if (!config) {
      return (
        <div className="h-[100px] flex items-center justify-center">
          No wireguard config found.
        </div>
      );
    }
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
        {loading ? (
          <LoadingPlaceHolder
            height={100}
            title={
              mode === 'config'
                ? 'Loading wireguard config...'
                : 'Loading wireguard config qr code...'
            }
          />
        ) : (
          modeView()
        )}
      </Popup.Content>
      {!loading && config && (
        <Popup.Footer>
          <Popup.Button
            onClick={() => {
              downloadConfig({
                filename: `${data.device}-wireguardconfig.yaml`,
                data: config,
              });
            }}
            content="Export"
            prefix={<ArrowLineDown />}
            variant="primary"
          />
        </Popup.Footer>
      )}
    </Popup.Root>
  );
};

type IDialog = IDialogBase<ExtractNodeType<IDevices>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();
  const params = useParams();

  const { cluster } = params;

  const { account } = useOutletContext<IAccountContext>();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: parseName(props.data),
            ports: props.data.spec?.ports || [],
          }
        : {
            displayName: '',
            name: '',
            ports: [],
          },
      validationSchema: Yup.object({
        name: Yup.string().required(),
        displayName: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        try {
          ensureClusterClientSide(params);

          if (!isUpdate) {
            const { errors } = await api.createVpnDevice({
              clusterName: ensureResource(cluster),
              vpnDevice: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                  namespace: ENV_NAMESPACE,
                },
                spec: {
                  accountName: parseName(account),
                  clusterName: ensureResource(cluster),
                  ports: val.ports,
                },
              },
            });
            if (errors) {
              throw errors[0];
            }
          } else if (isUpdate && props.data) {
            const { errors } = await api.updateVpnDevice({
              clusterName: cluster || '',
              vpnDevice: {
                displayName: val.displayName,
                metadata: {
                  name: parseName(props.data),
                  namespace: ENV_NAMESPACE,
                },
                spec: {
                  accountName: parseName(account),
                  clusterName: ensureResource(cluster),
                  ports: val.ports,
                },
              },
            });
            if (errors) {
              throw errors[0];
            }
          }

          reloadPage();
          resetValues();
          toast.success('Credential created successfully');
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-3xl">
          <div className="flex flex-col">
            <div className="flex flex-col gap-2xl">
              {isUpdate && (
                <Chips.Chip
                  {...{
                    item: { id: parseName(props.data) },
                    label: parseName(props.data),
                    prefix: 'Id:',
                    disabled: true,
                    type: 'BASIC',
                  }}
                />
              )}
              <TextInput
                label="Name"
                value={values.displayName}
                onChange={handleChange('displayName')}
                error={!!errors.displayName}
                message={errors.displayName}
              />
            </div>
            {!isUpdate && (
              <IdSelector
                resType="vpn_device"
                name={values.displayName}
                onChange={(value) =>
                  handleChange('name')({ target: { value } })
                }
                className="pt-2xl"
              />
            )}
          </div>
          <ExposedPorts
            ports={values.ports}
            onChange={(ports) => {
              handleChange('ports')(dummyEvent(ports));
            }}
          />
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

const HandleDevices = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Add device"
      updateTitle="Edit device"
      root={Root}
    />
  );
};

export default HandleDevices;
