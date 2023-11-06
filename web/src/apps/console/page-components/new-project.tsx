import {
  ArrowLeft,
  ArrowRight,
  CircleDashed,
  CircleFill,
  Search,
  UserCircle,
} from '@jengaicons/react';
import { useLoaderData, useNavigate, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Radio from '~/components/atoms/radio';
import { dayjs } from '~/components/molecule/dayjs';
import { toast } from '~/components/molecule/toast';
import { useMapper } from '~/components/utils';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import AlertModal from '../components/alert-modal';
import { IdSelector } from '../components/id-selector';
import NoResultsFound from '../components/no-results-found';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
import { SearchBox } from '../components/search-box';
import { FadeIn } from '../routes/_.$account.$cluster.$project.$scope.$workspace.new-app/util';
import { INewProjectFromAccountLoader } from '../routes/_a.$a.new-project';
import { parseName, parseNodes } from '../server/r-utils/common';
import { keyconstants } from '../server/r-utils/key-constants';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';

const NewProject = () => {
  const { cluster: clusterName } = useParams();
  const isOnboarding = !!clusterName;
  const { clustersData } = useLoaderData<INewProjectFromAccountLoader>();
  const clusters = parseNodes(clustersData);

  const api = useAPIClient();
  const navigate = useNavigate();

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);

  const { a: accountName } = useParams();

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      clusterName: isOnboarding ? clusterName : clusters[0]?.metadata?.name,
      nodeType: '',
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
          project: {
            metadata: {
              name: val.name,
              annotations: {
                [keyconstants.displayName]: val.displayName,
                [keyconstants.nodeType]: val.nodeType,
              },
            },
            displayName: val.displayName,
            spec: {
              clusterName: val.clusterName,
              accountName,
              targetNamespace: val.name,
            },
          },
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
        handleError(err);
      }
    },
  });

  const items = useMapper(
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
        ],
    (i) => {
      return {
        value: i.id,
        item: {
          ...i,
        },
      };
    }
  );

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
        progressItems={items}
        badge={{
          title: 'Kloudlite Labs Pvt Ltd',
          subtitle: accountName,
          image: <UserCircle size={20} />,
        }}
        onCancel={() => setShowUnsavedChanges(true)}
        rightChildren={
          <FadeIn onSubmit={handleSubmit}>
            <TitleBox
              title="Configure project"
              subtitle="Set up project settings and preferences."
            />
            <div className="flex flex-col">
              <TextInput
                label="Project name"
                name="name"
                value={values.displayName}
                onChange={handleChange('displayName')}
                size="lg"
                error={!!errors.displayName}
                message={errors.displayName}
              />
              <IdSelector
                className="pt-lg"
                resType="project"
                name={values.displayName}
                onChange={(v) => {
                  handleChange('name')(dummyEvent(v));
                }}
              />
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
                {clusters.length > 0 && (
                  <Radio.Root
                    withBounceEffect={false}
                    value={values.clusterName}
                    onChange={(e) => {
                      handleChange('clusterName')(dummyEvent(e));
                    }}
                    className="flex flex-col pr-2xl !gap-y-0 min-h-[288px]"
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
                              <div className="flex items-center flex-row flex-1 gap-lg">
                                <span className="headingMd text-text-default">
                                  {c.displayName}
                                </span>
                                <span>
                                  <CircleFill size={2} />
                                </span>
                                <span className="bodySm text-text-soft">
                                  {parseName(c)}
                                </span>
                              </div>
                              <span className="bodyMd text-text-default ">
                                {dayjs(c.updateTime).fromNow()}
                              </span>
                            </div>
                          </div>
                        </Radio.Item>
                      );
                    })}
                  </Radio.Root>
                )}
                {clusters.length === 0 && (
                  <NoResultsFound
                    title="No search results found."
                    image={<Search size={40} />}
                    border={false}
                    shadow={false}
                  />
                )}
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
          </FadeIn>
        }
      />

      <AlertModal
        title="Leave page with unsaved changes?"
        message="Leaving this page will delete all unsaved changes."
        okText="Leave page"
        cancelText="Stay"
        variant="critical"
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
