import useDebounce from '~/root/lib/client/hooks/use-debounce';
import * as Popover from '~/components/molecule/popover';
import * as Chips from '~/components/atoms/chips';
import { ChangeEvent, useEffect, useState } from 'react';
import { PencilLine } from '@jengaicons/react';
import { TextInput } from '~/components/atoms/input';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { useParams } from '@remix-run/react';
import { NonNullableString } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { ConsoleResType, ResType } from '~/root/src/generated/gql/server';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';

// export const idTypes = {
//   app: 'app',
//   project: 'project',
//   secret: 'secret',
//   config: 'config',
//   router: 'router',
//   managedresource: 'managedresource',
//   managedservice: 'managedservice',
//   workspace: 'workspace',
//   environment: 'environment',
//
//   cluster: 'cluster',
//
//   providersecret: 'providersecret',
//   nodepool: 'nodepool',
//   account: 'account',
// };

interface IidSelector {
  name: string;
  resType: ConsoleResType | ResType | 'account' | NonNullableString;
  onChange?: (id: string) => void;
}

export const IdSelector = ({
  name,
  onChange = (_) => {},
  resType,
}: IidSelector) => {
  const [id, setId] = useState(`my-awesome-${resType}`);
  const [idDisabled, setIdDisabled] = useState(true);
  const [popupId, setPopupId] = useState(id);
  const [isPopupIdValid, setPopupIdValid] = useState(true);
  const [idLoading, setIdLoading] = useState(false);
  const [popupOpen, setPopupOpen] = useState(false);

  useEffect(() => {
    if (name) {
      onChange(id);
    }
  }, [id]);

  const api = useAPIClient();
  const params = useParams();
  const { project, cluster } = params;

  const checkApi = (() => {
    switch (resType) {
      case 'app':
      case 'project':
      case 'config':
      case 'environment':
      case 'managedresource':
      case 'managedservice':
      case 'router':
      case 'secret':
      case 'workspace':
        ensureAccountClientSide(params);
        ensureClusterClientSide(params);
        return api.coreCheckNameAvailability;

      case 'cluster':
      case 'providersecret':
      case 'nodepool':
        ensureAccountClientSide(params);
        return api.infraCheckNameAvailability;

      case 'account':
        return api.accountCheckNameAvailability;

      default:
        return api.coreCheckNameAvailability;
    }
  })();

  console.log('params', params);

  useDebounce(
    async () => {
      if (name) {
        setIdLoading(true);
        try {
          const { data, errors } = await checkApi({
            resType,
            name: `${name}`,
            ...(project
              ? {
                  namespace: project,
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
        }
      }
    },
    500,
    [popupId]
  );

  const onPopupIdChange = ({ target }: ChangeEvent<HTMLInputElement>) => {
    setPopupIdValid(false);
    setPopupId(target.value);
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
    <Popover.Popover onOpenChange={setPopupOpen}>
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
          />
        </Popover.Footer>
      </Popover.Content>
    </Popover.Popover>
  );
};
