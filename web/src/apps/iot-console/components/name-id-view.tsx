/* eslint-disable no-nested-ternary */
import { ReactNode, forwardRef, useEffect, useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { NonNullableString } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { ConsoleResType, ResType } from '~/root/src/generated/gql/server';
import { useParams } from '@remix-run/react';
import { dummyEvent } from '~/root/lib/client/hooks/use-form';
import { CircleNotch } from '~/iotconsole/components/icons';
import { ensureAccountClientSide } from '../server/utils/auth-utils';
import { useIotConsoleApi } from '../server/gql/api-provider';

interface INameIdView {
  name: string;
  displayName: string;
  resType:
    | ConsoleResType
    | ResType
    | 'account'
    | 'username'
    | 'console_vpn_device'
    | NonNullableString;
  onChange?: ({ name, id }: { name: string; id: string }) => void;
  prefix?: ReactNode;
  errors?: string;
  label?: ReactNode;
  placeholder?: string;
  onCheckError?: (error: boolean) => void;
  isUpdate?: boolean;
  handleChange?: (key: string) => (e: {
    target: {
      value: string;
    };
  }) => void;
  nameErrorLabel: string;
}

export const NameIdView = forwardRef<HTMLInputElement, INameIdView>(
  (
    {
      name,
      onChange = (_) => {},
      resType,
      errors,
      prefix,
      label,
      displayName,
      placeholder,
      onCheckError,
      isUpdate,
      handleChange,
      nameErrorLabel,
    },
    ref
  ) => {
    const [nameValid, setNameValid] = useState(false);
    const [nameLoading, setNameLoading] = useState(true);

    const api = useIotConsoleApi();
    const params = useParams();

    const checkApi = (() => {
      switch (resType) {
        case 'app':
        case 'project':
        case 'config':
        case 'environment':
        case 'managed_service':
        case 'project_managed_service':
        case 'managed_resource':
        case 'router':
        case 'console_vpn_device':
        case 'secret':
          ensureAccountClientSide(params);
          return api.coreCheckNameAvailability;

        case 'cluster':
        case 'providersecret':
          ensureAccountClientSide(params);
          return api.infraCheckNameAvailability;
        case 'helm_release':
        case 'vpn_device':
        case 'nodepool':
          return api.infraCheckNameAvailability;

        case 'account':
          return api.accountCheckNameAvailability;

        case 'username':
          return api.crCheckNameAvailability;

        default:
          return api.coreCheckNameAvailability;
      }
    })();

    useEffect(() => {
      if (displayName && name) {
        setNameLoading(true);
      }
    }, [displayName, name]);

    const checkNameAvailable = () => {
      if (errors) {
        // onCheckError?.(true);
        return errors;
      }
      if (!name) {
        // onCheckError?.(true);
        return null;
      }

      if (isUpdate) {
        return null;
      }

      if (nameLoading) {
        // handleChange?.(nameErrorLabel)(dummyEvent(true));
        // setNameCheckError(true);
        return (
          <div className="flex flex-row items-center gap-md">
            <span className="animate-spin">
              <CircleNotch size={10} />
            </span>
            <span>Checking availability</span>
          </div>
        );
      }
      if (nameValid) {
        // handleChange?.(nameErrorLabel)(dummyEvent(false));
        // setNameCheckError(false);
        return (
          <span className="text-text-success bodySm-semibold">
            {name} is available.
          </span>
        );
      }
      const error = 'This name is not available. Please try different.';
      onCheckError?.(!!error);
      // handleChange?.(nameErrorLabel)(dummyEvent(!!error));
      // setNameCheckError(!!error);
      return error;
    };

    const { cluster, environment } = params;
    useDebounce(
      async () => {
        let tempResType = resType;
        if (resType === 'console_vpn_device') {
          tempResType = 'vpn_device';
        }
        if (!isUpdate)
          if (displayName) {
            setNameLoading(true);
            handleChange?.(nameErrorLabel)(dummyEvent(true));
            try {
              // @ts-ignore
              const { data, errors } = await checkApi({
                // @ts-ignore
                resType: tempResType,
                name: `${name}`,
                ...([
                  'project',
                  'app',
                  'environment',
                  'config',
                  'secret',
                  'project_managed_service',
                  'console_vpn_device',
                  'router',
                ].includes(tempResType)
                  ? {
                      
                      envName: environment,
                    }
                  : {}),
                ...(['nodepool', 'vpn_device', 'helm_release'].includes(
                  tempResType
                )
                  ? {
                      clusterName: cluster,
                    }
                  : {}),
                ...(tempResType === 'managed_resource'
                  ? {
                      namespace: '',
                    }
                  : {}),
              });

              if (errors) {
                throw errors[0];
              }
              if (data.result) {
                setNameValid(true);
                handleChange?.(nameErrorLabel)(dummyEvent(false));
              } else {
                setNameValid(false);
                handleChange?.(nameErrorLabel)(dummyEvent(true));
              }
            } catch (err) {
              handleError(err);
            } finally {
              setNameLoading(false);
            }
          } else {
            setNameLoading(false);
          }
      },
      500,
      [displayName, name, isUpdate]
    );

    return (
      <TextInput
        ref={ref}
        label={label}
        value={displayName}
        onChange={(e) => {
          const v = e.target.value;
          const id = v.trim().toLowerCase().replace(/ /g, '-');
          onChange?.({
            name: v,
            id,
          });
          handleChange?.('displayName')(dummyEvent(v));
          if (!isUpdate) {
            handleChange?.('name')(dummyEvent(id));
          }
          if (v) {
            setNameLoading(true);
            if (!isUpdate) {
              handleChange?.(nameErrorLabel)(dummyEvent(true));
            }
          } else {
            setNameLoading(false);
            handleChange?.(nameErrorLabel)(dummyEvent(false));
          }
        }}
        placeholder={placeholder}
        size="lg"
        error={
          (!nameLoading || !isUpdate) &&
          ((!nameValid && !!name && !nameLoading) || !!errors)
        }
        message={checkNameAvailable()}
        prefix={
          prefix && <span className="text-text-soft mr-sm">{prefix} /</span>
        }
        focusRing
      />
    );
  }
);
