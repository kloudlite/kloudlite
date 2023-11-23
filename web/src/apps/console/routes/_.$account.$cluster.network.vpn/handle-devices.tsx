import {
  ArrowLineDown,
  ArrowRight,
  Check,
  ChevronLeft,
  ChevronRight,
  Plus,
  SmileySad,
  X,
} from '@jengaicons/react';
import { useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import Chips from '~/components/atoms/chips';
import { NumberInput, TextInput } from '~/components/atoms/input';
import { usePagination } from '~/components/molecule/pagination';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { cn } from '~/components/utils';
import { IdSelector } from '~/console/components/id-selector';
import List from '~/console/components/list';
import NoResultsFound from '~/console/components/no-results-found';
import QRCodeView from '~/console/components/qr-code';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IDevices } from '~/console/server/gql/queries/vpn-queries';
import {
  ExtractNodeType,
  ensureResource,
  parseName,
} from '~/console/server/r-utils/common';
import { ensureClusterClientSide } from '~/console/server/utils/auth-utils';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import { ENV_NAMESPACE } from '~/root/lib/configs/env';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import {
  InfoLabel,
  parseValue,
} from '../_.$account.$cluster.$project.$scope.$workspace.new-app/util';

interface IExposedPorts {
  targetPort?: number;
  port: number;
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

export const ShowQR = ({ show, setShow }: IDialog<string>) => {
  return (
    <Popup.Root
      show={show as any}
      onOpenChange={(e) => {
        setShow(e);
      }}
    >
      <Popup.Header>QR Code</Popup.Header>
      <Popup.Content>
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
            <QRCodeView value={show?.data || 'Error'} />
          </div>
        </div>
      </Popup.Content>
    </Popup.Root>
  );
};

export const ShowWireguardConfig = ({ show, setShow }: IDialog) => {
  return (
    <Popup.Root show={show as any} onOpenChange={setShow}>
      <Popup.Header>WireGuard Config</Popup.Header>
      <Popup.Content>
        <div className="flex flex-col gap-3xl">
          <div className="bodyMd text-text-default">
            Please use the following configuration to set up your WireGuard
            client.
          </div>
          <div className="p-3xl flex flex-col gap-3xl border border-border-default rounded-lg">
            <div className="pb-3xl flex flex-col gap-lg">
              <div className="bodyMd-medium text-text-soft">Interface</div>
              <div className="flex flex-col gap-md text-text-default">
                <div className="flex flex-row gap-4xl ">
                  <span className="bodyMd-medium w-9xl">PrivateKey</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">YJGz9Lk/80Q</span>
                </div>
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">Address</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">10.6.0.2/32</span>
                </div>
              </div>
            </div>
            <div className="flex flex-col gap-lg">
              <div className="bodyMd-medium text-text-soft">Peer</div>
              <div className="flex flex-col gap-md text-text-default">
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">PublicKey</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">Yy4QH9ik6vbl</span>
                </div>
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">AllowedIPs</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">0.0.0.0/0</span>
                </div>
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">Endpoint</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">PersistentKeepalive/25</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button
          content="Export"
          prefix={<ArrowLineDown />}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

const HandleDevices = ({
  show,
  setShow,
}: IDialog<ExtractNodeType<IDevices> | null>) => {
  const api = useConsoleApi();
  const reloadPage = useReload();
  const params = useParams();

  const [validationSchema, setValidationSchema] = useState(
    Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
    })
  );

  const { cluster } = params;

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
      displayName: '',
      name: '',
      ports: [],
    },
    validationSchema,
    onSubmit: async (val) => {
      try {
        ensureClusterClientSide(params);

        if (show?.type === DIALOG_TYPE.ADD) {
          const { errors } = await api.createVpnDevice({
            clusterName: ensureResource(cluster),
            vpnDevice: {
              displayName: val.displayName,
              metadata: {
                name: val.name,
                namespace: ENV_NAMESPACE,
              },
              spec: {
                serverName: 'server',
                ports: val.ports,
              },
            },
          });
          if (errors) {
            throw errors[0];
          }
        } else if (show?.data) {
          const { errors } = await api.updateVpnDevice({
            clusterName: cluster || '',
            vpnDevice: {
              displayName: val.displayName,
              metadata: {
                name: parseName(show.data),
                namespace: ENV_NAMESPACE,
              },
              spec: {
                serverName: 'server',
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
        setShow(null);
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (show && show.data && show.type === DIALOG_TYPE.EDIT) {
      setValues((v) => ({
        ...v,
        displayName: show.data?.displayName || '',
      }));
      setValidationSchema(
        // @ts-ignore
        Yup.object({
          displayName: Yup.string().trim().required(),
        })
      );
    }
  }, [show]);
  return (
    <Popup.Root
      show={show as any}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>
        {show?.type === DIALOG_TYPE.ADD ? 'Add new device' : 'Edit device'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-3xl">
            <div className="flex flex-col">
              <div className="flex flex-col gap-2xl">
                {show?.type === DIALOG_TYPE.EDIT && (
                  <Chips.Chip
                    {...{
                      item: { id: parseName(show.data) },
                      label: parseName(show.data),
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
              {show?.type === DIALOG_TYPE.ADD && (
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
            content={show?.type === DIALOG_TYPE.ADD ? 'Create' : 'Update'}
            variant="primary"
            loading={isLoading}
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleDevices;
