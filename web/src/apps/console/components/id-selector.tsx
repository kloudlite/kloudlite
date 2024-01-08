import { PencilLine } from '@jengaicons/react';
import { useOutletContext, useParams } from '@remix-run/react';
import { ChangeEvent, useEffect, useState } from 'react';
import AnimateHide from '~/components/atoms/animate-hide';
import Chips from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import Popover from '~/components/molecule/popover';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { NonNullableString } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { ConsoleResType, ResType } from '~/root/src/generated/gql/server';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';
import { IEnvironmentContext } from '../routes/_main+/$account+/$project+/$environment+/_layout';

interface IidSelector {
  name: string;
  resType:
    | ConsoleResType
    | ResType
    | 'account'
    | 'username'
    | NonNullableString;
  onChange?: (id: string) => void;
  onLoad?: (loading: boolean) => void;
  className?: string;
}

export const IdSelector = ({
  name,
  onChange = (_) => {},
  onLoad = (_) => {},
  resType,
  className,
}: IidSelector) => {
  const [id, setId] = useState('');
  const [idDisabled, setIdDisabled] = useState(true);
  const [popupId, setPopupId] = useState(id);
  const [isPopupIdValid, setPopupIdValid] = useState(true);
  const [idLoading, setIdLoading] = useState(false);
  const [idInternalLoading, setIdInternalLoading] = useState(false);
  const [popupOpen, setPopupOpen] = useState(false);

  useEffect(() => {
    if (name) {
      onChange(id);
    }
  }, [id]);

  useEffect(() => {
    onLoad(idLoading);
  }, [idLoading]);

  const api = useAPIClient();
  const params = useParams();
  const { cluster } = params;
  const { environment, project } = useOutletContext<IEnvironmentContext>();

  const checkApi = (() => {
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
  })();

  useDebounce(
    async () => {
      if (name) {
        setIdLoading(true);
        try {
          const { data, errors } = await checkApi({
            resType,
            name: `${name}`,
            // eslint-disable-next-line no-nested-ternary
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
          // console.log(data, errors);
          if (data.result) {
            setId(`${name}`);
            setPopupId(`${name}`);
          } else if (data.suggestedNames.length > 0) {
            setId(data.suggestedNames[0]);
            setPopupId(data.suggestedNames[0]);
          }
          setIdDisabled(false);
        } catch (err) {
          handleError(err);
        } finally {
          setIdLoading(false);
        }
      } else {
        setIdDisabled(true);
      }
    },
    500,
    [name]
  );

  useDebounce(
    async () => {
      if (popupId && popupOpen) {
        try {
          const { data, errors } = await checkApi({
            resType,
            name: `${popupId}`,
          });

          if (errors) {
            throw errors[0];
          }
          if (data.result) {
            setPopupIdValid(true);
          } else {
            setPopupIdValid(false);
          }
        } catch (err) {
          handleError(err);
        } finally {
          setIdInternalLoading(false);
        }
      } else {
        setIdInternalLoading(false);
      }
    },
    500,
    [popupId]
  );

  const onPopupIdChange = ({ target }: ChangeEvent<HTMLInputElement>) => {
    setPopupIdValid(false);
    setPopupId(target.value);
    if (target.value) {
      setIdInternalLoading(true);
    }
  };

  const onPopupCancel = () => {
    setPopupId(id);
  };

  const onPopupSave = () => {
    setId(popupId);
  };

  useEffect(() => {
    if (name === '') {
      setIdDisabled(true);
    }
  }, [name]);

  return (
    <AnimateHide show={!!id && !!name}>
      <div className={className}>
        <Popover.Root onOpenChange={setPopupOpen}>
          <Popover.Trigger>
            <Chips.Chip
              label={id}
              prefix={<PencilLine />}
              type="CLICKABLE"
              loading={idLoading}
              disabled={idDisabled}
              item={{ clusterId: id }}
            />
          </Popover.Trigger>
          <Popover.Content align="start">
            <TextInput
              label={
                <span>
                  <span className="capitalize">{resType}</span> ID
                </span>
              }
              error={!idInternalLoading && !isPopupIdValid}
              message={
                !idInternalLoading && !isPopupIdValid
                  ? 'This id is not available. Please try different.'
                  : null
              }
              value={popupId}
              onChange={onPopupIdChange}
            />
            <Popover.Footer>
              <Popover.Button
                variant="basic"
                content="Cancel"
                size="sm"
                onClick={onPopupCancel}
              />
              <Popover.Button
                variant="primary"
                content="Save"
                size="sm"
                type="button"
                disabled={!isPopupIdValid}
                onClick={onPopupSave}
                loading={idInternalLoading}
              />
            </Popover.Footer>
          </Popover.Content>
        </Popover.Root>
      </div>
    </AnimateHide>
  );
};
