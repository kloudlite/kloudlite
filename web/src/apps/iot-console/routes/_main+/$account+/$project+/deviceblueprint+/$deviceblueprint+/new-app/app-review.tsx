import { useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { useAppState } from '~/iotconsole/page-components/app-states';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import useForm from '~/lib/client/hooks/use-form';
import Yup from '~/lib/server/helpers/yup';
import { handleError } from '~/lib/utils/common';
import { validateType } from '~/root/src/generated/gql/validator';
import { parseName } from '~/iotconsole/server/r-utils/common';
import { FadeIn } from '~/iotconsole/page-components/util';
import {
  BottomNavigation,
  ReviewComponent,
} from '~/iotconsole/components/commons';
import { keyconstants } from '~/iotconsole/server/r-utils/key-constants';
import { IDeviceBlueprintContext } from '../_layout';

const AppReview = () => {
  const { app, setPage, resetState } = useAppState();

  const api = useIotConsoleApi();
  const navigate = useNavigate();
  const { project, deviceblueprint } =
    useOutletContext<IDeviceBlueprintContext>();
  const [projectName, deviceblueprintName] = [
    project.name,
    deviceblueprint.name,
    // parseName(account),
  ];

  const { handleSubmit, isLoading } = useForm({
    initialValues: app,
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      if (!project || !deviceblueprint) {
        throw new Error('Project and deviceblueprint is required!.');
      }

      try {
        const { errors } = await api.createIotApp({
          projectName,
          deviceBlueprintName: deviceblueprintName,
          app: {
            ...app,
          },
        });
        if (errors) {
          throw errors[0];
        }

        toast.success('App created successfully');

        resetState();
        navigate('../apps');
      } catch (err) {
        handleError(err);
      }
    },
  });

  const [errors, setErrors] = useState<string[]>([]);

  useEffect(() => {
    const res = validateType(app, 'AppIn');
    setErrors(res);
  }, []);

  return (
    <FadeIn onSubmit={handleSubmit}>
      <div className="bodyMd text-text-soft">
        An assessment of the work, product, or performance.
      </div>
      <div className="flex flex-col gap-3xl">
        <ReviewComponent
          title="Application detail"
          onEdit={() => {
            setPage(1);
          }}
        >
          <div className="flex flex-col rounded border border-border-default">
            <div className="flex flex-col p-xl gap-md">
              <div className="bodyMd-semibold text-text-default">
                {app.displayName}
              </div>
              <div className="bodySm text-text-soft">{parseName(app)}</div>
            </div>
          </div>
        </ReviewComponent>

        <ReviewComponent
          title="Compute"
          onEdit={() => {
            setPage(2);
          }}
        >
          <div className="flex flex-row gap-3xl">
            <div className="flex flex-col rounded border border-border-default flex-1 overflow-hidden">
              <div className="px-xl py-lg bg-surface-basic-subdued">
                Container image
              </div>
              {app.spec.containers.map((container) => {
                return (
                  <div
                    key={container.name}
                    className="p-xl flex flex-col gap-md"
                  >
                    <div className="bodyMd-medium text-text-default">
                      {container.image}
                    </div>
                    <div className="bodySm text-text-soft">
                      {container.name}
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="flex flex-col rounded border border-border-default flex-1 overflow-hidden">
              <div className="px-xl py-lg bg-surface-basic-subdued">
                Plan details
              </div>
              <div className="p-xl flex flex-col gap-md">
                <div className="bodyMd-medium text-text-default">
                  Essential plan
                </div>
                <div className="bodySm text-text-soft">0.35vCPU & 0.35GB</div>
              </div>
            </div>
          </div>
        </ReviewComponent>

        {!!app.spec.nodeSelector?.[keyconstants.nodepoolName] && (
          <ReviewComponent title="Nodepool Details" onEdit={() => {}}>
            <div className="flex flex-col p-xl gap-md rounded border border-border-default">
              <div className="bodyMd-semibold text-text-default">
                Nodepool Selector
              </div>
              <div className="bodySm text-text-soft">
                {app.spec.nodeSelector?.[keyconstants.nodepoolName]}
              </div>
            </div>
          </ReviewComponent>
        )}

        <ReviewComponent
          title="Environment"
          onEdit={() => {
            setPage(3);
          }}
        >
          <div className="flex flex-col gap-xl p-xl rounded border border-border-default">
            <div className="flex flex-row items-center gap-lg pb-xl border-b border-border-default">
              <div className="flex-1 bodyMd-medium text-text-default">
                Environment variables
              </div>
              <div className="text-text-soft bodyMd">
                {app.spec.containers[0].env?.length || 0}
              </div>
            </div>
            <div className="flex flex-row items-center gap-lg">
              <div className="flex-1 bodyMd-medium text-text-default">
                Config mount
              </div>
              <div className="text-text-soft bodyMd">
                {app.spec.containers[0].volumes?.length || 0}
              </div>
            </div>
          </div>
        </ReviewComponent>
        <ReviewComponent
          title="Network"
          onEdit={() => {
            setPage(4);
          }}
        >
          <div className="flex flex-row gap-xl p-xl rounded border border-border-default">
            <div className="text-text-default bodyMd flex-1">
              Ports exposed from the app
            </div>
            <div className="text-text-soft bodyMd">
              {app.spec.services?.length || 0}
            </div>
          </div>
        </ReviewComponent>
      </div>

      {errors.length > 0 && (
        <div className="text-text-critical flex flex-col gap-md">
          <span className="font-bold">Errors:</span>
          {errors.map((i, index) => {
            return (
              <pre key={i}>
                {index + 1}. {i}
              </pre>
            );
          })}
        </div>
      )}

      <BottomNavigation
        primaryButton={{
          type: 'submit',
          content: 'Create App',
          variant: 'primary',
          disabled: errors.length !== 0,
          loading: isLoading,
        }}
        secondaryButton={{
          content: 'Network',
          variant: 'outline',
          onClick: () => {
            setPage(4);
          },
        }}
      />
    </FadeIn>
  );
};

export default AppReview;
