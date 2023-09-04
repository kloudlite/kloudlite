import { ArrowLeft, ArrowRight, CircleDashed } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { useState } from 'react';
import {
  useLoaderData,
  useNavigate,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import Radio from '~/components/atoms/radio';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '~/components/molecule/toast';
import { dayjs } from '~/components/molecule/dayjs';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { Badge } from '~/components/atoms/badge';
import { cn } from '~/components/utils';
import ProgressTracker from '~/components/organisms/progress-tracker';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';
import {
  getMetadata,
  parseDisplaynameFromAnn,
  parseName,
  parseUpdationTime,
} from '../server/r-urils/common';
import { IdSelector } from '../components/id-selector';
import { SearchBox } from '../components/search-box';
import { getProject, getProjectSepc } from '../server/r-urils/project';
import { keyconstants } from '../server/r-urils/key-constants';
import RawWrapper from '../components/raw-wrapper';
import AlertDialog from '../components/alert-dialog';

const NewProject = () => {
  const { cluster: clusterName } = useParams();
  const isOnboarding = !!clusterName;
  const { clustersData, cluster } = useLoaderData();
  const clusters = clustersData?.edges?.map(({ node }) => node || []);

  const api = useAPIClient();
  const navigate = useNavigate();

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);

  const { user, account } = useOutletContext();
  const { a: accountName } = useParams();

  const { values, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      clusterName: isOnboarding ? clusterName : '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      clusterName: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      try {
        ensureClusterClientSide({ cluster: val.clusterName });
        ensureAccountClientSide({ account: accountName });
        const { errors: e } = await api.createProject({
          project: getProject({
            metadata: getMetadata({
              name: val.name,
              annotations: {
                [keyconstants.displayName]: val.displayName,
                [keyconstants.author]: user.name,
                [keyconstants.node_type]: val.node_type,
              },
            }),
            displayName: val.displayName,
            spec: getProjectSepc({
              clusterName: val.clusterName,
              accountName,
              targetNamespace: val.name,
            }),
          }),
        });

        if (e) {
          throw e[0];
        }
        toast.success('project added successfully');
        navigate(
          isOnboarding
            ? `/onboarding/${accountName}/${val.name}/congratulations`
            : '/projects'
        );
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  return (
    <>
      <RawWrapper
        title={
          isOnboarding
            ? 'Begin Your Project Journey.'
            : 'Let’s create new project.'
        }
        subtitle={
          isOnboarding
            ? 'Kloudlite will help you to develop and deploy cloud native applications easily.'
            : 'Create your project under production effortlessly.'
        }
        progressItems={
          isOnboarding
            ? [
                {
                  label: 'Create Team',
                  active: true,
                  id: 1,
                  completed: false,
                },
                {
                  label: 'Invite your Team Members',
                  active: true,
                  id: 2,
                  completed: false,
                },
                {
                  label: 'Add your Cloud Provider',
                  active: true,
                  id: 3,
                  completed: false,
                },
                {
                  label: 'Setup First Cluster',
                  active: true,
                  id: 4,
                  completed: false,
                },
                {
                  label: 'Create your project',
                  active: true,
                  id: 5,
                  completed: false,
                },
              ]
            : [
                {
                  label: 'Configure project',
                  active: true,
                  id: 1,
                  completed: false,
                },
                {
                  label: 'Review',
                  active: false,
                  id: 2,
                  completed: false,
                },
              ]
        }
        rightChildren={
          <form onSubmit={handleSubmit} className="gap-6xl flex flex-col">
            <div className="text-text-soft headingLg">Configure projects</div>
            <div className="flex flex-col gap-4xl">
              <div className="flex flex-col gap-3xl">
                <TextInput
                  label="Project name"
                  name="name"
                  value={values.displayName}
                  onChange={handleChange('displayName')}
                  size="lg"
                />
                <IdSelector
                  resType="project"
                  name={values.displayName}
                  onChange={(v) => {
                    handleChange('name')(dummyEvent(v));
                  }}
                />
              </div>
            </div>
            {!isOnboarding && (
              <div className="flex flex-col border border-border-disabled bg-surface-basic-default rounded-md">
                <div className="bg-surface-basic-subdued flex flex-row items-center py-lg pr-lg pl-2xl">
                  <span className="headingMd text-text-default flex-1">
                    Cluster(s)
                  </span>
                  <div className="flex-1">
                    <SearchBox InputElement={TextInput} />
                  </div>
                </div>
                <Radio.Root
                  value={values.clusterName}
                  onChange={(e) => {
                    handleChange('clusterName')(dummyEvent(e));
                  }}
                  className="flex flex-col pr-2xl !gap-y-0"
                  labelPlacement="left"
                >
                  {clusters?.map((c) => {
                    return (
                      <Radio.Item
                        value={parseName(c)}
                        withBounceEffect={false}
                        className="justify-between w-full"
                        key={parseName(c)}
                      >
                        <div className="p-2xl pl-lg flex flex-row gap-lg items-center">
                          <CircleDashed size={24} />
                          <div className="flex flex-row flex-1 items-center gap-lg">
                            <span className="headingMd text-text-default">
                              {parseName(c)}
                            </span>
                            <span className="bodyMd text-text-default ">
                              {dayjs(parseUpdationTime(c)).fromNow()}
                            </span>
                          </div>
                        </div>
                      </Radio.Item>
                    );
                  })}
                </Radio.Root>
              </div>
            )}

            {isOnboarding ? (
              <div className="flex flex-row gap-xl justify-end">
                <Button
                  variant="outline"
                  content="Back"
                  prefix={<ArrowLeft />}
                  size="lg"
                />
                <Button
                  loading={isLoading}
                  variant="primary"
                  content="Get started"
                  suffix={<ArrowRight />}
                  size="lg"
                  type="submit"
                />
              </div>
            ) : (
              <div className="flex flex-row justify-end">
                <Button
                  loading={isLoading}
                  variant="primary"
                  content="Create"
                  suffix={<ArrowRight />}
                  type="submit"
                  size="lg"
                />
              </div>
            )}
          </form>
        }
      />

      <AlertDialog
        title="Leave page with unsaved changes?"
        message="Leaving this page will delete all unsaved changes."
        okText="Leave page"
        type="critical"
        show={showUnsavedChanges}
        setShow={setShowUnsavedChanges}
        onSubmit={() => {
          setShowUnsavedChanges(false);
          navigate(`/${accountName}/projects`);
        }}
      />
    </>
  );
};

export default NewProject;
