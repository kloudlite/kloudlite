import { ArrowLeft, ArrowRight, CircleDashed } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
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
import logger from '~/root/lib/client/helpers/log';
import { dayjs } from '~/components/molecule/dayjs';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { Badge } from '~/components/atoms/badge';
import { cn } from '~/components/utils';
import {
  ensureAccountClientSide,
  ensureAccountSet,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';
import {
  getMetadata,
  getPagination,
  getSearch,
  parseName,
  parseUpdationTime,
} from '../server/r-urils/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { IdSelector } from '../components/id-selector';
import { SearchBox } from '../components/search-box';
import { getProject, getProjectSepc } from '../server/r-urils/project';
import { keyconstants } from '../server/r-urils/key-constants';
import RawWrapper from '../components/raw-wrapper';
import AlertDialog from '../components/alert-dialog';

const NewProject = () => {
  const [isOnboarding, setIsOnboarding] = useState(false);
  const { clustersData } = useLoaderData();
  const clusters = clustersData?.edges?.map(({ node }) => node || []);

  const api = useAPIClient();
  const navigate = useNavigate();

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);

  // @ts-ignore
  const { user } = useOutletContext();
  const { a: account } = useParams();

  const { values, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      clusterName: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      clusterName: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      try {
        ensureClusterClientSide({ cluster: val.clusterName });
        ensureAccountClientSide({ account });
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
            spec: getProjectSepc({
              clusterName: val.clusterName,
              displayName: val.displayName,
              accountName: account,
              targetNamespace: val.name,
            }),
          }),
        });

        if (e) {
          throw e[0];
        }
        toast.success('project added successfully');
        navigate('/projects');
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  return (
    <>
      <RawWrapper
        leftChildren={
          <>
            <BrandLogo detailed={false} size={48} />
            <div
              className={cn('flex flex-col', {
                'gap-4xl': isOnboarding,
                'gap-8xl': !isOnboarding,
              })}
            >
              <div className="flex flex-col gap-3xl">
                <div className="text-text-default heading4xl">
                  {isOnboarding
                    ? 'Begin Your Project Journey.'
                    : 'Let’s create new project.'}
                </div>
                <div className="text-text-default bodyMd">
                  {isOnboarding
                    ? 'Kloudlite will help you to develop and deploy cloud native applications easily.'
                    : 'Create your project under production effortlessly.'}
                </div>
                {!isOnboarding && (
                  <div className="flex flex-row gap-md items-center">
                    <Badge>
                      <span className="text-text-strong">Account:</span>
                      <span className="bodySm-semibold text-text-default">
                        Kloudlite Labs Pvt Ltd
                      </span>
                    </Badge>
                  </div>
                )}
              </div>
              <ProgressTracker
                items={
                  isOnboarding
                    ? [
                        { label: 'Create Team', active: true, id: 1 },
                        {
                          label: 'Invite your Team Members',
                          active: true,
                          id: 2,
                        },
                        {
                          label: 'Add your Cloud Provider',
                          active: true,
                          id: 3,
                        },
                        { label: 'Setup First Cluster', active: true, id: 4 },
                        { label: 'Create your project', active: true, id: 5 },
                      ]
                    : [
                        { label: 'Configure project', active: true, id: 1 },
                        { label: 'Review', active: false, id: 2 },
                      ]
                }
              />
            </div>
            {isOnboarding && (
              <Button variant="outline" content="Skip" size="lg" />
            )}
            {!isOnboarding && (
              <Button
                variant="outline"
                content="Cancel"
                size="lg"
                onClick={() => setShowUnsavedChanges(true)}
              />
            )}
          </>
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
                  name={values.displayName}
                  onChange={(v) => {
                    handleChange('name')(dummyEvent(v));
                  }}
                />
              </div>
            </div>
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
                {clusters?.map((cluster) => {
                  return (
                    <Radio.Item
                      value={parseName(cluster)}
                      withBounceEffect={false}
                      className="justify-between w-full"
                      key={parseName(cluster)}
                    >
                      <div className="p-2xl pl-lg flex flex-row gap-lg items-center">
                        <CircleDashed size={24} />
                        <div className="flex flex-row flex-1 items-center gap-lg">
                          <span className="headingMd text-text-default">
                            {parseName(cluster)}
                          </span>
                          <span className="bodyMd text-text-default ">
                            {dayjs(parseUpdationTime(cluster)).fromNow()}
                          </span>
                        </div>
                      </div>
                    </Radio.Item>
                  );
                })}
              </Radio.Root>
            </div>
            {isOnboarding ? (
              <div className="flex flex-row gap-xl justify-end">
                <Button
                  variant="outline"
                  content="Back"
                  prefix={ArrowLeft}
                  size="lg"
                />
                <Button
                  variant="primary"
                  content="Get started"
                  suffix={ArrowRight}
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
                  suffix={ArrowRight}
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
          navigate(`/${account}/projects`);
        }}
      />
    </>
  );
};

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  const { data, errors } = await GQLServerHandler(ctx.request).listClusters({
    pagination: getPagination(ctx),
    search: getSearch(ctx),
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    clustersData: data,
  };
};

export default NewProject;
