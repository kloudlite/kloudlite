/* eslint-disable no-nested-ternary */
import { ArrowRight } from '@jengaicons/react';
import { useLoaderData, useNavigate, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { toast } from '~/components/molecule/toast';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Select from '~/components/atoms/select';
import { parseName, parseNodes } from '../server/r-utils/common';
import { keyconstants } from '../server/r-utils/key-constants';
import {
  ensureAccountClientSide,
  ensureClusterClientSide,
} from '../server/utils/auth-utils';
import { INewProjectFromAccountLoader } from '../routes/_a+/$a+/new-project';
import ProgressWrapper from '../components/progress-wrapper';
import { useConsoleApi } from '../server/gql/api-provider';
import { NameIdView } from '../components/name-id-view';
import { ReviewComponent } from '../routes/_main+/$account+/$project+/$environment+/new-app/app-review';

type steps = 'Configure project' | 'Review';
const NewProject = () => {
  const { cluster: clusterName } = useParams();
  const isOnboarding = !!clusterName;
  const { clustersData } = useLoaderData<INewProjectFromAccountLoader>();
  const clusters = parseNodes(clustersData);

  const [activeState, setActiveState] = useState<steps>('Configure project');
  const isActive = (step: steps) => step === activeState;

  const api = useConsoleApi();
  const navigate = useNavigate();

  const params = useParams();
  const { a: accountName } = params;

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      clusterName: isOnboarding ? clusterName : clusters[0]?.metadata?.name,
      nodeType: '',
      isNameError: false,
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      clusterName: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      const submit = async () => {
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
              clusterName: val.clusterName,
              displayName: val.displayName,
              spec: {
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
      };

      switch (activeState) {
        case 'Configure project':
          setActiveState('Review');
          break;
        case 'Review':
          await submit();
          break;
        default:
          break;
      }
    },
  });

  const getView = () => {
    return (
      <form
        className="flex flex-col gap-3xl py-3xl"
        onSubmit={(e) => {
          if (!values.isNameError) {
            handleSubmit(e);
          } else {
            e.preventDefault();
          }
        }}
      >
        <div className="bodyMd text-text-soft">
          Create your project under production effortlessly.
        </div>

        <NameIdView
          label="Project name"
          resType="project"
          name={values.name}
          displayName={values.displayName}
          errors={errors.name}
          prefix={accountName}
          handleChange={handleChange}
          nameErrorLabel="isNameError"
        />

        {!isOnboarding && (
          <Select
            label="Cluster"
            placeholder="Select a cluster"
            value={{
              label:
                clusters.find((c) => parseName(c) === values.clusterName)
                  ?.displayName || values.clusterName,
              value: values.clusterName,
            }}
            options={async () => [
              ...clusters.map((clster) => ({
                label: clster.displayName,
                value: parseName(clster),
                cluster: clster,
                render: () => (
                  <div className="flex flex-col">
                    <div>{clster.displayName}</div>
                    <div className="bodySm text-text-soft">
                      {parseName(clster)}
                    </div>
                  </div>
                ),
              })),
            ]}
            onChange={(v) => {
              handleChange('clusterName')(dummyEvent(v.value));
            }}
          />
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

  const getReviewView = () => {
    return (
      <form onSubmit={handleSubmit} className="flex flex-col gap-3xl py-3xl">
        <ReviewComponent
          title="Project detail"
          onEdit={() => {
            setActiveState('Configure project');
          }}
        >
          <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
            <div className="flex flex-col gap-md  pb-lg  border-b border-border-default">
              <div className="bodyMd-semibold text-text-default">
                Project name
              </div>
              <div className="bodySm text-text-soft">{values.name}</div>
            </div>
            <div className="flex flex-col gap-md  pb-lg  border-b border-border-default">
              <div className="bodyMd-semibold text-text-default">Cluster</div>
              <div className="bodySm text-text-soft">{values.clusterName}</div>
            </div>
          </div>
        </ReviewComponent>
        <div className="flex flex-row justify-start">
          <Button
            loading={isLoading}
            variant="primary"
            content="Create"
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
        active: isActive('Configure project'),
        completed: false,
        children: isActive('Configure project') ? getView() : null,
      },
      {
        label: 'Review',
        active: isActive('Review'),
        completed: false,
        children: isActive('Review') ? getReviewView() : null,
      },
    ];
  };

  return (
    <ProgressWrapper
      title={isOnboarding ? 'Setup your account!' : 'Let’s create new project.'}
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite teams"
      progressItems={{
        items: getItems(),
      }}
      onClick={() => {
        if (!isOnboarding) {
          if (isActive('Review')) {
            setActiveState('Configure project');
          }
        }
      }}
    />
  );
};

export default NewProject;
