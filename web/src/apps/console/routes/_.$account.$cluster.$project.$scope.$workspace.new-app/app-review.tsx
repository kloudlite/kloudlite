import { PencilLine } from '@jengaicons/react';
import { ReactNode } from 'react';

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
  return (
    <>
      <div className="flex flex-col gap-xl">
        <div className="headingXl text-text-default">Review</div>
        <div className="bodySm text-text-soft">
          An assessment of the work, product, or performance.
        </div>
      </div>
      <div className="flex flex-col gap-3xl">
        <ReviewComponent title="Application detail" onEdit={() => {}}>
          <div className="flex flex-col p-xl gap-md rounded border border-border-default">
            <div className="bodyMd-semibold text-text-default">Audrey</div>
            <div className="bodySm text-text-soft">Audrey-1234590ng</div>
          </div>
        </ReviewComponent>

        <ReviewComponent title="Compute" onEdit={() => {}}>
          <div className="flex flex-row gap-3xl">
            <div className="flex flex-col rounded border border-border-default flex-1 overflow-hidden">
              <div className="px-xl py-lg bg-surface-basic-subdued">
                Container image
              </div>
              <div className="p-xl flex flex-col gap-md">
                <div className="bodyMd-medium text-text-default">
                  /github.com/projects/123img
                </div>
                <div className="bodySm text-text-soft">ezord_aws_key_id</div>
              </div>
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
              <div className="text-text-soft bodyMd">08</div>
            </div>
            <div className="flex flex-row items-center gap-lg">
              <div className="flex-1 bodyMd-medium text-text-default">
                Config mount
              </div>
              <div className="text-text-soft bodyMd">06</div>
            </div>
          </div>
        </ReviewComponent>
        <ReviewComponent title="Network" onEdit={() => {}}>
          <div className="flex flex-row gap-xl p-xl rounded border border-border-default">
            <div className="text-text-default bodyMd">Total no. of network</div>
            <div className="text-text-soft bodyMd">06</div>
          </div>
        </ReviewComponent>
      </div>
    </>
  );
};

export default AppReview;
