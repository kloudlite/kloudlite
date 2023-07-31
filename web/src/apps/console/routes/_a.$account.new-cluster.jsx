import { ArrowLeft, ArrowRight, PencilLine } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import { useState } from 'react';
import { useParams } from '@remix-run/react';
import * as SelectInput from '~/components/atoms/select';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from 'react-toastify';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { getCookie } from '~/root/lib/app-setup/cookies';
import * as Popover from '~/components/molecule/popover';
import * as Tooltip from '~/components/atoms/tooltip';
import * as Chips from '~/components/atoms/chips';
import { useAPIClient } from '../server/utils/api-provider';

const NewCluster = () => {
  const api = useAPIClient();
  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);
  const [clusterId, setClusterId] = useState('my-awesome-cluster');
  const [clusterIdDisabled, setClusterIdDisabled] = useState(true);
  const [popupClusterId, setPopupClusterId] = useState(clusterId);
  const [isPopupClusterIdValid, setPopupClusterIdValid] = useState(true);
  const [clusterIdLoading, setClusterIdLoading] = useState(false);
  const [popupOpen, setPopupOpen] = useState(false);

  const cookie = getCookie();
  const { account } = useParams();

  const { values, handleSubmit, handleChange } = useForm({
    initialValues: {},
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      try {
        console.log(values);
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  useDebounce(values.name, 500, async () => {
    if (values.name) {
      setClusterIdLoading(true);
      try {
        cookie.set('kloudlite-account', account);
        const { data, errors } = await api.checkNameAvailability({
          resType: 'cluster',
          name: `${values.name}-cluster`,
        });

        if (errors) {
          throw errors[0];
        }
        if (data.result) {
          setClusterId(`${values.name}-cluster`);
          setPopupClusterId(`${values.name}-cluster`);
        } else if (data.suggestedNames.length > 0) {
          setClusterId(data.suggestedNames[0]);
          setPopupClusterId(data.suggestedNames[0]);
        }
        setClusterIdDisabled(false);
      } catch (err) {
        toast.error(err.message);
      } finally {
        setClusterIdLoading(false);
      }
    } else {
      setClusterIdDisabled(true);
    }
  });

  useDebounce(popupClusterId, 500, async () => {
    if (popupClusterId && popupOpen) {
      try {
        cookie.set('kloudlite-account', account);
        const { data, errors } = await api.checkNameAvailability({
          resType: 'cluster',
          name: `${popupClusterId}`,
        });

        if (errors) {
          throw errors[0];
        }
        if (data.result) {
          setPopupClusterIdValid(true);
        } else {
          setPopupClusterIdValid(false);
        }
      } catch (err) {
        toast.error(err.message);
      }
    }
  });

  const onClusterIdChange = (e) => {
    if (e.target.value === '') {
      setClusterIdDisabled(true);
    }
    handleChange('name')(e);
  };

  const onPopupClusterIdChange = ({ target }) => {
    setPopupClusterIdValid(false);
    setPopupClusterId(target.value);
  };

  const onPopupCancel = () => {
    setPopupClusterId(clusterId);
  };

  const onPopupSave = () => {
    setClusterId(popupClusterId);
  };

  return (
    <Tooltip.TooltipProvider>
      <div className="h-full flex flex-row">
        <div className="h-full w-[571px] flex flex-col bg-surface-basic-subdued py-11xl px-10xl">
          <div className="flex flex-col gap-8xl">
            <div className="flex flex-col gap-4xl items-start">
              <BrandLogo detailed={false} size={48} />
              <div className="flex flex-col gap-3xl">
                <div className="text-text-default heading4xl">
                  Let’s create new cluster.
                </div>
                <div className="text-text-default bodyLg">
                  Create your cluster to production effortlessly
                </div>
              </div>
            </div>
            <ProgressTracker
              items={[
                { label: 'Configure cluster', active: true, id: 1 },
                { label: 'Review', active: false, id: 2 },
              ]}
            />
            <Button
              variant="outline"
              content="Back"
              prefix={ArrowLeft}
              onClick={() => setShowUnsavedChanges(true)}
            />
          </div>
        </div>
        <form className="py-11xl px-10xl flex-1" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4xl">
            <div className="h-7xl" />
            <div className="flex flex-col gap-3xl p-3xl">
              <TextInput
                label="Cluster name"
                name="name"
                onChange={onClusterIdChange}
                value={values.name}
              />
              <Popover.Popover onOpenChange={setPopupOpen}>
                <Popover.Trigger>
                  <Chips.Chip
                    label={clusterId}
                    prefix={PencilLine}
                    type={Chips.ChipType.CLICKABLE}
                    loading={clusterIdLoading}
                    disabled={clusterIdDisabled}
                    item={{ clusterId }}
                  />
                </Popover.Trigger>
                <Popover.Content align="start">
                  <TextInput
                    label="Cluster ID"
                    name="popupClusterId"
                    value={popupClusterId}
                    onChange={onPopupClusterIdChange}
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
                      disabled={!isPopupClusterIdValid}
                      onClick={onPopupSave}
                    />
                  </Popover.Footer>
                </Popover.Content>
              </Popover.Popover>
              <SelectInput.Select label="Provider" value="aws">
                <SelectInput.Option>AWS</SelectInput.Option>
              </SelectInput.Select>
              <SelectInput.Select label="Region" value="india">
                <SelectInput.Option>India</SelectInput.Option>
              </SelectInput.Select>
            </div>
          </div>
          <div className="flex flex-row justify-end px-3xl">
            <Button
              variant="primary"
              content="Create"
              suffix={ArrowRight}
              type="submit"
            />
          </div>
        </form>

        {/* Unsaved change alert dialog */}
        <AlertDialog.DialogRoot
          show={showUnsavedChanges}
          onOpenChange={setShowUnsavedChanges}
        >
          <AlertDialog.Header>
            Leave page with unsaved changes?
          </AlertDialog.Header>
          <AlertDialog.Content>
            Leaving this page will delete all unsaved changes.
          </AlertDialog.Content>
          <AlertDialog.Footer>
            <AlertDialog.Button variant="basic" content="Cancel" />
            <AlertDialog.Button variant="critical" content="Delete" />
          </AlertDialog.Footer>
        </AlertDialog.DialogRoot>
      </div>
    </Tooltip.TooltipProvider>
  );
};

export default NewCluster;
