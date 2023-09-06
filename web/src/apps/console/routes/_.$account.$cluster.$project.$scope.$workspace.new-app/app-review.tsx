import { ArrowLeft, ArrowRight, PencilLine } from '@jengaicons/react';
import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { FadeIn } from './util';
import { useAppState } from './states';

interface IReviewComponent {
  title: string;
  children: ReactNode;
  onEdit: () => void;
}
const ReviewComponent = ({
  title = '',
  children,
  onEdit,
}: IReviewComponent) => {
  return (
    <div className="flex flex-col gap-2xl pb-3xl">
      <div className="flex flex-row items-center">
        <span className="text-text-soft bodyMd flex-1">{title}</span>
        <span className="text-icon-soft" onClick={onEdit}>
          <PencilLine size={16} />
        </span>
      </div>
      {children}
    </div>
  );
};
const AppReview = () => {
  const { app, setPage } = useAppState();

  const api = useConsoleApi();

  const { handleSubmit, isLoading } = useForm({
    initialValues: app,
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      try {
        const { errors } = await api.createApp({
          app,
        });
        if (errors) {
          throw errors[0];
        }
        toast.success('created successfully');
      } catch (err) {
        handleError(err);
      }
    },
  });

  return (
    <FadeIn onSubmit={handleSubmit}>
      <div className="flex flex-col gap-xl">
        <div className="headingXl text-text-default">Review</div>
        <div className="bodyMd text-text-soft">
          An assessment of the work, product, or performance.
        </div>
      </div>
      <div className="flex flex-col gap-3xl">
        <ReviewComponent title="Application detail" onEdit={() => {}}>
          <div className="flex flex-col p-xl gap-md rounded border border-border-default">
            <div className="bodyMd-semibold text-text-default">
              {app.displayName}
            </div>
            <div className="bodySm text-text-soft">{app.metadata.name}</div>
          </div>
        </ReviewComponent>

        <ReviewComponent title="Compute" onEdit={() => {}}>
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
        <ReviewComponent title="Environment" onEdit={() => {}}>
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
        <ReviewComponent title="Network" onEdit={() => {}}>
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
      <div className="flex flex-row gap-xl justify-end items-center">
        <Button
          content="Networks"
          prefix={<ArrowLeft />}
          variant="outline"
          onClick={() => {
            setPage('network');
          }}
        />

        <div className="text-surface-primary-subdued">|</div>

        <Button
          content="Create App"
          suffix={<ArrowRight />}
          variant="primary"
          type="submit"
          loading={isLoading}
        />
      </div>
    </FadeIn>
  );
};

export default AppReview;
