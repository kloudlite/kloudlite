/* eslint-disable no-nested-ternary */
import { ArrowRight } from '@jengaicons/react';
import { useLoaderData, useNavigate, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { toast } from '~/components/molecule/toast';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Select from '~/components/atoms/select';
import { listStatus, parseStatus } from '~/console/components/sync-status';
import { IClusters } from '~/console/server/gql/queries/cluster-queries';
import * as cluster from 'cluster';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '../server/r-utils/common';
import { ensureAccountClientSide } from '../server/utils/auth-utils';
import { INewProjectFromAccountLoader } from '../routes/_a+/$a+/new-project';
import { useConsoleApi } from '../server/gql/api-provider';
import { NameIdView } from '../components/name-id-view';
import { ReviewComponent } from '../routes/_main+/$account+/$project+/$environment+/new-app/app-review';
import MultiStepProgress, {
  useMultiStepProgress,
} from '../components/multi-step-progress';
import MultiStepProgressWrapper from '../components/multi-step-progress-wrapper';
import { TitleBox } from '../components/raw-wrapper';

const statusRender = (item: ExtractNodeType<IClusters>) => {
  return listStatus({
    key: parseName(item),
    item,
    className: 'text-center',
  });
};

const NewProject = () => {
  const { clustersData } = useLoaderData<INewProjectFromAccountLoader>();
  const clusters = parseNodes(clustersData);

  const api = useConsoleApi();
  const navigate = useNavigate();

  const params = useParams();
  const { a: accountName } = params;
  const rootUrl = `/${accountName}/projects`;

  const { currentStep, jumpStep, nextStep } = useMultiStepProgress({
    defaultStep: 1,
    totalSteps: 2,
  });

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      clusterName: '',
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
          ensureAccountClientSide({ account: accountName });
          const { errors: e } = await api.createProject({
            project: {
              metadata: {
                name: val.name,
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
          navigate(rootUrl);
        } catch (err) {
          handleError(err);
        }
      };

      switch (currentStep) {
        case 1:
          nextStep();
          break;
        case 2:
          await submit();
          break;
        default:
          break;
      }
    },
  });

  return (
    <form
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <MultiStepProgressWrapper
        title="Let’s create new project."
        subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite teams"
        backButton={{
          content: 'Back to projects',
          to: rootUrl,
        }}
      >
        <MultiStepProgress.Root currentStep={currentStep} jumpStep={jumpStep}>
          <MultiStepProgress.Step step={1} label="Configure project">
            <div className="flex flex-col gap-3xl">
              <TitleBox subtitle="Create your project under production effortlessly." />
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
                  ...clusters
                    .filter(
                      (clster) => parseStatus({ item: clster }) !== 'deleting'
                    )
                    .map((clster) => ({
                      label: clster.displayName,
                      value: parseName(clster),
                      cluster: clster,
                      render: () => (
                        <div>
                          {parseStatus({ item: clster }) === 'ready' ? (
                            <div className="flex flex-col">
                              <div>{clster.displayName}</div>
                              <div className="bodySm text-text-soft">
                                {parseName(clster)}
                              </div>
                            </div>
                          ) : (
                            // parseStatus({ item: clster }) === 'syncing' ||
                            // parseStatus({ item: clster }) === 'notready' ?
                            <div className="flex text-text-disabled">
                              <div className="flex-grow">
                                <div className="flex flex-col">
                                  <div>{clster.displayName}</div>
                                  <div className="bodySm">
                                    {parseName(clster)}
                                  </div>
                                </div>
                              </div>
                              <div className="flex flex-grow-0 items-center">
                                {statusRender(
                                  clster as ExtractNodeType<IClusters>
                                ).render()}
                              </div>
                            </div>
                          )}
                        </div>
                      ),
                    })),
                ]}
                onChange={(v) => {
                  if (parseStatus({ item: v.cluster }) === 'ready')
                    handleChange('clusterName')(dummyEvent(v.value));
                }}
              />
              <div className="flex flex-row justify-start">
                <Button
                  variant="primary"
                  content="Next"
                  suffix={<ArrowRight />}
                  type="submit"
                />
              </div>
            </div>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step step={2} label="Review">
            <ReviewComponent
              title="Project detail"
              onEdit={() => {
                jumpStep(1);
              }}
            >
              <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
                <div className="flex flex-col gap-md  pb-lg  border-b border-border-default">
                  <div className="bodyMd-semibold text-text-default">
                    Project name
                  </div>
                  <div className="bodySm text-text-soft">{values.name}</div>
                </div>
                <div className="flex flex-col gap-md">
                  <div className="bodyMd-semibold text-text-default">
                    Cluster
                  </div>
                  <div className="bodySm text-text-soft">
                    {values.clusterName}
                  </div>
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
          </MultiStepProgress.Step>
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </form>
  );
};

export default NewProject;
