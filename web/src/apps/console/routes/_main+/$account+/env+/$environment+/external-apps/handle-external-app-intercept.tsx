/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
import Popup from '~/components/molecule/popup';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useCallback, useEffect, useState } from 'react';
import Select from '~/components/atoms/select';
import { mapper } from '~/components/utils';
import { useOutletContext } from '@remix-run/react';
import { ISetState } from '~/console/page-components/app-states';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { IExternalApps } from '~/console/server/gql/queries/external-app-queries';
import { IEnvironmentContext } from '../_layout';
import ExposedExternalAppPortList, {
  exposedExternalAppPortsType,
} from './external-app-network';
// import ExposedPortList, { exposedPortsType } from './network';

type IDialog = {
  app?: ExtractNodeType<IExternalApps>;
  visible: boolean;
  setVisible: ISetState<boolean>;
};

const Root = (props: IDialog) => {
  const { visible, setVisible, app } = props;

  const api = useConsoleApi();
  const { environment } = useOutletContext<IEnvironmentContext>();
  const { data: dData, isLoading: dIsLoading } = useCustomSwr(
    'devices',
    async () =>
      api.listGlobalVpnDevices({
        gvpn: 'default',
        pagination: {
          first: 100,
        },
      }),
    true
  );

  const devices = useCallback(() => parseNodes(dData), [dData])();

  const [ports, setPorts] = useState<exposedExternalAppPortsType[]>([]);

  useEffect(() => {
    if (app) {
      setPorts(
        app.spec?.intercept?.portMappings?.map((s) => {
          return {
            appPort: s.appPort,
            devicePort: s.devicePort,
          };
        }) || []
        // app.spec?.services?.map((s) => {
        //   return {
        //     appPort: s.port,
        //     devicePort:
        //       app.spec.intercept?.portMappings?.find(
        //         (v) => v.appPort === s.port
        //       )?.devicePort || s.port,
        //   };
        // }) || []
      );
    }
  }, [app]);

  const reloadPage = useReload();

  const {
    values,
    errors,
    handleSubmit,
    handleChange,
    isLoading,
    resetValues,
    setValues,
  } = useForm({
    initialValues: app
      ? {
          deviceName: app.spec?.intercept?.toDevice || '',
        }
      : {},
    validationSchema: Yup.object({
      deviceName: Yup.string().required(),
    }),

    onSubmit: async (val) => {
      if (!val.deviceName) {
        return;
      }

      const appName = parseName(app);
      if (!appName) {
        toast.error('app is not provided');
        return;
      }
      try {
        const { errors: e } = await api.interceptExternalApp({
          deviceName: val.deviceName,
          intercept: true,
          envName: parseName(environment),
          externalAppName: parseName(app),
          portMappings: ports,
        });
        if (e) {
          throw e[0];
        }

        reloadPage();
        setVisible(false);
        toast.success('External App Intercepted successfully');
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (devices.length) {
      setValues((v) => ({ ...v, deviceName: parseName(devices[0]) }));
    }
  }, [dData]);

  useEffect(() => {
    resetValues();
  }, [visible]);

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          <Select
            label="Select Device"
            size="lg"
            value={values.deviceName}
            disabled={dIsLoading}
            placeholder="select a device"
            options={async () =>
              mapper(devices, (d) => {
                return {
                  ...d,
                  value: parseName(d),
                  label: parseName(d),
                };
              })
            }
            onChange={(_, value) => {
              handleChange('deviceName')(dummyEvent(value));
            }}
            error={!!errors.clusterName}
            message={errors.clusterName}
            loading={dIsLoading}
          />

          <ExposedExternalAppPortList
            setExposedPorts={setPorts}
            exposedPorts={ports}
          />
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content="Intercept"
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleExternalAppIntercept = (props: IDialog) => {
  const { setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>Intercept App</Popup.Header>
      <Root {...props} />
    </Popup.Root>
  );
};

export default HandleExternalAppIntercept;
