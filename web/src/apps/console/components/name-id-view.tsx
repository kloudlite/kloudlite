/* eslint-disable no-nested-ternary */
import { CircleNotch } from '@jengaicons/react';
import { ReactNode, useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
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
}

export const NameIdView = ({
  name,
  onChange = (_) => {},
  resType,
  errors,
  prefix,
  label,
  displayName,
}: INameIdView) => {
  const [nameValid, setNameValid] = useState(false);
  const [nameLoading, setNameLoading] = useState(false);

  const api = useConsoleApi();
  const params = useParams();

  const checkApi = () => {
    switch (resType) {
      case 'app':
      case 'project':
      case 'config':
      case 'environment':
      case 'managed_service':
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
  };

  const checkNameAvailable = () => {
    if (errors) {
      return errors;
    }
    if (!name) {
      return null;
    }
    if (nameLoading) {
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
      return (
        <span className="text-text-success bodySm-semibold">
          {name} is available.
        </span>
      );
    }
    return 'This name is not available. Please try different.';
  };

  const { environment, project } = useOutletContext<IEnvironmentContext>();
  const { cluster } = params;
  useDebounce(
    async () => {
      if (displayName) {
        try {
          // @ts-ignore
          const { data, errors } = await checkApi()({
            resType,
            name: `${name}`,
            ...(resType === 'environment'
              ? {
                  namespace: project.spec?.targetNamespace,
                }
              : environment
              ? {
                  namespace: environment.spec?.targetNamespace,
                }
              : {}),
            ...(resType === 'nodepool' || resType === 'vpn_device'
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
    [displayName, name]
  );

  return (
    <TextInput
      label={label}
      value={displayName}
      onChange={(e) => {
        const v = e.target.value;
        onChange?.({
          name: v,
          id: v.toLowerCase().replace(' ', '-'),
        });
        if (v) {
          setNameLoading(true);
        } else {
          setNameLoading(false);
        }
      }}
      size="lg"
      error={!nameValid && !!name && !nameLoading}
      message={checkNameAvailable()}
      prefix={
        prefix && <span className="text-text-soft mr-sm">{prefix} /</span>
      }
    />
  );
};
