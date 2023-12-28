import {
  ArrowRight,
  CircleDashed,
  CircleFill,
  Search,
} from '@jengaicons/react';
import { useLoaderData, useNavigate, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Radio from '~/components/atoms/radio';
import { dayjs } from '~/components/molecule/dayjs';
import { toast } from '~/components/molecule/toast';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IdSelector } from '../components/id-selector';
import NoResultsFound from '../components/no-results-found';
import { SearchBox } from '../components/search-box';
import {
  parseName,
  parseNamespace,
  parseNodes,
  parseTargetNs,
} from '../server/r-utils/common';
import { keyconstants } from '../server/r-utils/key-constants';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';
import { INewProjectFromAccountLoader } from '../routes/_a+/$a+/new-project';
import ProgressWrapper from '../components/progress-wrapper';
import { useConsoleApi } from '../server/gql/api-provider';

const NewProject = () => {
  const { cluster: clusterName } = useParams();
  const isOnboarding = !!clusterName;
  const { clustersData } = useLoaderData<INewProjectFromAccountLoader>();
  const clusters = parseNodes(clustersData);

  const api = useConsoleApi();
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
            // spec: {
            //   clusterName: val.clusterName,
            //   accountName,
            //   targetNamespace: val.name,
            // },
            spec: {
              displayName: val.displayName,
              targetNamespace: '',
            },
          },
        });

        if (e) {
          throw e[0];
        }
        toast.success('project created successfully');
        navigate(`/${accountName}/projects`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const getView = () => {
    return (
      <form className="flex flex-col gap-3xl py-3xl" onSubmit={handleSubmit}>
        <div className="bodyMd text-text-soft">
          Create your project under production effortlessly.
        </div>
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
        <div className="flex flex-row justify-start">
          <Button
            loading={isLoading}
            variant="primary"
            content="Next"
            suffix={<ArrowRight />}
            type="submit"
          />
        </div>
      </form>
    );
  };

  const getItems = () => {
    return [
      {
        label: 'Configure project',
        active: true,
        completed: false,
        children: getView(),
      },
      {
        label: 'Review',
        active: false,
        completed: false,
      },
    ];
  };

  return (
    <ProgressWrapper
      title={isOnboarding ? 'Setup your account!' : 'Let’s create new project.'}
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite
  teams"
      progressItems={{
        items: getItems(),
      }}
    />
  );
};

export default NewProject;
