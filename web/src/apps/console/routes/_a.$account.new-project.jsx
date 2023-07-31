import {
  ArrowLeft,
  ArrowRight,
  CircleDashed,
  PencilLine,
  Search,
} from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import { useState } from 'react';
import { useParams } from '@remix-run/react';
import * as Radio from '~/components/atoms/radio';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from 'react-toastify';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { getCookie } from '~/root/lib/app-setup/cookies';
import * as Popover from '~/components/molecule/popover';
import * as Tooltip from '~/components/atoms/tooltip';
import * as Chips from '~/components/atoms/chips';
import { useAPIClient } from '../server/utils/api-provider';

const NewProject = () => {
  const api = useAPIClient();
  const [clusters, setClusters] = useState([
    {
      label: 'Plaxonic',
      time: '. 197d ago',
      id: 1,
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
      id: 2,
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
      id: 3,
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
      id: 4,
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
      id: 5,
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
      id: 6,
    },
  ]);

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);
  const [projectId, setProjectId] = useState('my-awesome-project');
  const [projectIdDisabled, setProjectIdDisabled] = useState(true);
  const [popupProjectId, setPopupProjectId] = useState(projectId);
  const [isPopupProjectIdValid, setPopupProjectIdValid] = useState(true);
  const [projectIdLoading, setProjectIdLoading] = useState(false);
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
      setProjectIdLoading(true);
      try {
        cookie.set('kloudlite-account', account);
        const { data, errors } = await api.checkNameAvailability({
          resType: 'cluster',
          name: `${values.name}-project`,
        });

        if (errors) {
          throw errors[0];
        }
        if (data.result) {
          setProjectId(`${values.name}-project`);
          setPopupProjectId(`${values.name}-project`);
        } else if (data.suggestedNames.length > 0) {
          setProjectId(data.suggestedNames[0]);
          setPopupProjectId(data.suggestedNames[0]);
        }
        setProjectIdDisabled(false);
      } catch (err) {
        toast.error(err.message);
      } finally {
        setProjectIdLoading(false);
      }
    } else {
      setProjectIdDisabled(true);
    }
  });

  useDebounce(popupProjectId, 500, async () => {
    if (popupProjectId && popupOpen) {
      try {
        cookie.set('kloudlite-account', account);
        const { data, errors } = await api.checkNameAvailability({
          resType: 'cluster',
          name: `${popupProjectId}`,
        });

        if (errors) {
          throw errors[0];
        }
        if (data.result) {
          setPopupProjectIdValid(true);
        } else {
          setPopupProjectIdValid(false);
        }
      } catch (err) {
        toast.error(err.message);
      }
    }
  });

  const onClusterIdChange = (e) => {
    if (e.target.value === '') {
      setProjectIdDisabled(true);
    }
    handleChange('name')(e);
  };

  const onPopupClusterIdChange = ({ target }) => {
    setPopupProjectIdValid(false);
    setPopupProjectId(target.value);
  };

  const onPopupCancel = () => {
    setPopupProjectId(projectId);
  };

  const onPopupSave = () => {
    setProjectId(popupProjectId);
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
                  Let’s create new project.
                </div>
                <div className="text-text-default bodyLg">
                  Create your project to production effortlessly
                </div>
              </div>
            </div>
            <ProgressTracker
              items={[
                { label: 'Configure project', active: true, id: 1 },
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
          <div className="gap-6xl flex flex-col p-3xl">
            <div className="flex flex-col gap-4xl">
              <div className="h-7xl" />
              <div className="flex flex-col gap-3xl">
                <TextInput
                  label="Project name"
                  name="name"
                  onChange={onClusterIdChange}
                  value={values.name}
                />
                <Popover.Popover onOpenChange={setPopupOpen}>
                  <Popover.Trigger>
                    <Chips.Chip
                      label={projectId}
                      prefix={PencilLine}
                      type={Chips.ChipType.CLICKABLE}
                      loading={projectIdLoading}
                      disabled={projectIdDisabled}
                      item={{ projectId }}
                    />
                  </Popover.Trigger>
                  <Popover.Content align="start">
                    <TextInput
                      label="Project ID"
                      name="popupProjectId"
                      value={popupProjectId}
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
                        disabled={!isPopupProjectIdValid}
                        onClick={onPopupSave}
                      />
                    </Popover.Footer>
                  </Popover.Content>
                </Popover.Popover>
              </div>
            </div>
            <div className="flex flex-col border border-border-disabled bg-surface-basic-default rounded-md">
              <TextInput
                prefixIcon={Search}
                placeholder="Cluster(s)"
                className="bg-surface-basic-subdued rounded-none rounded-t-md border-0 border-b border-border-disabled"
              />
              <Radio.RadioGroup
                className="flex flex-col pr-2xl !gap-y-0"
                labelPlacement="left"
              >
                {clusters.map((cluster) => {
                  return (
                    <Radio.RadioItem
                      value={cluster.id}
                      withBounceEffect={false}
                      className="justify-between w-full"
                      key={cluster.id}
                    >
                      <div className="p-2xl pl-lg flex flex-row gap-lg items-center">
                        <CircleDashed size={24} />
                        <div className="flex flex-row flex-1 items-center gap-lg">
                          <span className="headingMd text-text-default">
                            {cluster.label}
                          </span>
                          <span className="bodyMd text-text-default ">
                            {cluster.time}
                          </span>
                        </div>
                      </div>
                    </Radio.RadioItem>
                  );
                })}
              </Radio.RadioGroup>
            </div>
            <div className="flex flex-row justify-end">
              <Button
                variant="primary"
                content="Create"
                suffix={ArrowRight}
                type="submit"
              />
            </div>
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

export default NewProject;
