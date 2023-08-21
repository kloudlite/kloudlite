import useDebounce from '~/root/lib/client/hooks/use-debounce';
import * as Popover from '~/components/molecule/popover';
import * as Chips from '~/components/atoms/chips';
import { useEffect, useState } from 'react';
import { PencilLine } from '@jengaicons/react';
import { TextInput } from '~/components/atoms/input';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { toast } from '~/components/molecule/toast';
import { useParams } from '@remix-run/react';
import { uuid } from '~/components/utils';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';

export const idTypes = {
  app: 'app',
  project: 'project',
  secret: 'secret',
  config: 'config',
  router: 'router',
  managedresource: 'managedresource',
  managedservice: 'managedservice',
  workspace: 'workspace',

  cluster: 'cluster',

  providersecret: 'providersecret',
  nodepool: 'nodepool',
  account: 'account',
};

export const IdSelector = ({
  name,
  onChange = (_) => {},
  resType = 'cluster',
}) => {
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
  const { project } = params;

  const checkApi = (() => {
    switch (resType) {
      case idTypes.app:
      case idTypes.project:
      case idTypes.secret:
      case idTypes.config:
      case idTypes.router:
      case idTypes.managedresource:
      case idTypes.managedservice:
      case idTypes.workspace:
        ensureAccountClientSide(params);
        ensureClusterClientSide(params);
        return api.coreCheckNameAvailability;

      case idTypes.cluster:
      case idTypes.providersecret:
      case idTypes.nodepool:
        ensureAccountClientSide(params);
        return api.infraCheckNameAvailability;

      case idTypes.account:
        // TODO: replace with api when available
        return async ({ resType: _, name: n }) => {
          toast.info(
            'TODO: used dummy api, have to replace with actual, [checkNameAvailability]'
          );
          return {
            data: {
              result: false,
              suggestedNames: [
                `${n.replaceAll(' ', '-').toLowerCase()}-${uuid().substring(
                  0,
                  4
                )}`,
              ],
            },
          };
        };

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
            ...(project
              ? {
                  namespace: project,
                }
              : {}),
          });

          if (errors) {
            throw errors[0];
          }
          console.log(data, errors);
          if (data.result) {
            setId(`${name}`);
            setPopupId(`${name}`);
          } else if (data.suggestedNames.length > 0) {
            setId(data.suggestedNames[0]);
            setPopupId(data.suggestedNames[0]);
          }
          setIdDisabled(false);
        } catch (err) {
          toast.error(err.message);
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
          toast.error(err.message);
        }
      }
    },
    500,
    [popupId]
  );

  const onPopupIdChange = ({ target }) => {
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
          prefix={PencilLine}
          type={Chips.ChipType.CLICKABLE}
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
