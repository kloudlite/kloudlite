import { ArrowLineDown } from '@jengaicons/react';
import { useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import Chips from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IdSelector } from '~/console/components/id-selector';
import QRCodeView from '~/console/components/qr-code';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IDevices } from '~/console/server/gql/queries/vpn-queries';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { ensureClusterClientSide } from '~/console/server/utils/auth-utils';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import { ENV_NAMESPACE } from '~/root/lib/configs/env';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

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
    },
    validationSchema,
    onSubmit: async (val) => {
      try {
        ensureClusterClientSide(params);

        if (show?.type === DIALOG_TYPE.ADD) {
          const { errors } = await api.createVpnDevice({
            vpnDevice: {
              displayName: val.displayName,
              metadata: {
                name: val.name,
                namespace: ENV_NAMESPACE,
              },
            },
          });
          if (errors) {
            throw errors[0];
          }
        } else if (show?.data) {
          const { errors } = await api.updateVpnDevice({
            vpnDevice: {
              displayName: val.displayName,
              metadata: {
                name: parseName(show.data),
                namespace: ENV_NAMESPACE,
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
