/* eslint-disable no-nested-ternary */
import { CircleNotch } from '@jengaicons/react';
import { ReactNode, useEffect, useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { NonNullableString } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { ConsoleResType, ResType } from '~/root/src/generated/gql/server';
import { useOutletContext, useParams } from '@remix-run/react';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';
import { useConsoleApi } from '../server/gql/api-provider';
import { IEnvironmentContext } from '../routes/_main+/$account+/$project+/$environment+/_layout';
import { parseName } from '../server/r-utils/common';

interface INameIdView {
  name: string;
  displayName: string;
  resType:
    | ConsoleResType
    | ResType
    | 'account'
    | 'username'
    | NonNullableString;
  onChange?: ({ name, id }: { name: string; id: string }) => void;
  prefix?: ReactNode;
  errors?: string;
  label?: ReactNode;
  placeholder?: string;
  onCheckError?: (error: boolean) => void;
  isUpdate?: boolean;
}

export const NameIdView = ({
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
}: INameIdView) => {
  const [nameValid, setNameValid] = useState(false);
  const [nameLoading, setNameLoading] = useState(true);

  const api = useConsoleApi();
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
      case 'helm_release':
      case 'router':
      case 'secret':
        ensureAccountClientSide(params);
        ensureClusterClientSide(params);
        return api.coreCheckNameAvailability;

      case 'cluster':
      case 'providersecret':
        ensureAccountClientSide(params);
        return api.infraCheckNameAvailability;
      case 'vpn_device':
        ensureClusterClientSide(params);
        ensureAccountClientSide(params);
        return api.infraCheckNameAvailability;
      case 'nodepool':
        ensureAccountClientSide(params);
        ensureClusterClientSide(params);
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
      onCheckError?.(true);
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
      onCheckError?.(false);
      return (
        <span className="text-text-success bodySm-semibold">
          {name} is available.
        </span>
      );
    }
    const error = 'This name is not available. Please try different.';
    onCheckError?.(!!error);
    return error;
  };

  const { environment, project } = useOutletContext<IEnvironmentContext>();
  const { cluster } = params;
  useDebounce(
    async () => {
      if (!isUpdate)
        if (displayName) {
          setNameLoading(true);
          try {
            // @ts-ignore
            const { data, errors } = await checkApi({
              // @ts-ignore
              resType,
              name: `${name}`,
              ...([
                'project',
                'app',
                'environment',
                'config',
                'secret',
                'project_managed_service',
              ].includes(resType)
                ? {
                    projectName: parseName(project),
                    envName: parseName(environment),
                  }
                : {}),
              ...(['nodepool', 'vpn_device'].includes(resType)
                ? {
                    clusterName: cluster,
                  }
                : {}),
              ...(resType === 'managed_resource'
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
            } else {
              setNameValid(false);
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

  useEffect(() => {
    console.log(
      'error: ',
      (!nameLoading || !isUpdate) &&
        ((!nameValid && !!name && !nameLoading) || !!errors),
      !nameLoading || !isUpdate,
      !nameValid && !!name && !nameLoading,
      !!errors
    );
  }, []);

  return (
    <TextInput
      label={label}
      value={displayName}
      onChange={(e) => {
        const v = e.target.value;
        onChange?.({
          name: v,
          id: v.trim().toLowerCase().replace(/ /g, '-'),
        });
        if (v) {
          setNameLoading(true);
        } else {
          setNameLoading(false);
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
    />
  );
};
