/* eslint-disable no-nested-ternary */
import { ArrowRight, CircleNotch } from '@jengaicons/react';
import { useLoaderData, useNavigate, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
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

const NewProject = () => {
  const { cluster: clusterName } = useParams();
  const isOnboarding = !!clusterName;
  const { clustersData } = useLoaderData<INewProjectFromAccountLoader>();
  const clusters = parseNodes(clustersData);

  const api = useConsoleApi();
  const navigate = useNavigate();

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);

  const params = useParams();
  const { a: accountName } = params;

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
    },
  });

  const getView = () => {
    return (
      <form className="flex flex-col gap-3xl py-3xl" onSubmit={handleSubmit}>
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
          onChange={({ name, id }) => {
            handleChange('displayName')(dummyEvent(name));
            handleChange('name')(dummyEvent(id));
          }}
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
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite teams"
      progressItems={{
        items: getItems(),
      }}
    />
  );
};

export default NewProject;
