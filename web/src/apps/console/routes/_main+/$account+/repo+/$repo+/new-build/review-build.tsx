import {
  GitBranchFill,
  GithubLogoFill,
  GitlabLogoFill,
} from '@jengaicons/react';
import { ReviewComponent } from '~/console/components/commons';

interface IReviewBuild {
  name: string;
  source: {
    branch: string;
    repository: string;
    provider: string;
  };
  tags: Array<string>;
  buildClusterName: string;
  advanceOptions: boolean;
  repository: string;
  buildArgs: Record<string, string>;
  buildContexts: Record<string, string>;
  contextDir: string;
  dockerfilePath: string;
  dockerfileContent: string;
}

const ReviewBuild = ({
  values,
  onEdit,
}: {
  values: IReviewBuild;
  onEdit: (step: number) => void;
}) => {
  const gitIconSize = 20;
  return (
    <div>
      <ReviewComponent title="Source details" onEdit={() => onEdit(1)}>
        <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
          <div className="flex flex-col gap-md">
            <div className="bodyMd-medium text-text-default">Source</div>
            <div className="flex flex-row items-center gap-3xl bodySm">
              <div className="flex flex-row items-center gap-xl">
                {values.source.provider === 'github' ? (
                  <GithubLogoFill size={gitIconSize} />
                ) : (
                  <GitlabLogoFill size={gitIconSize} />
                )}
                <span>
                  {values.source.repository
                    .replace('https://', '')
                    .replace('.git', '')}
                </span>
              </div>
              <div className="flex flex-row items-center gap-xl">
                <GitBranchFill size={16} />
                <span>{values.source.branch}</span>
              </div>
            </div>
          </div>
        </div>
      </ReviewComponent>
      <ReviewComponent title="Build details" onEdit={() => onEdit(2)}>
        <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
          <div className="flex flex-col gap-md  pb-lg border-b border-border-default">
            <div className="bodyMd-medium text-text-default">Build name</div>
            <div className="bodySm text-text-soft">{values.name}</div>
          </div>
          <div className="flex flex-col gap-md">
            <div className="bodyMd-medium text-text-default">Build Cluster</div>
            <div className="bodySm text-text-soft">
              {values.buildClusterName}
            </div>
          </div>
        </div>
      </ReviewComponent>
      <ReviewComponent title="Tags" onEdit={() => onEdit(2)}>
        <div className="flex flex-col gap-xl p-xl rounded border border-border-default">
          <div className="flex flex-row items-center gap-lg">
            <div className="flex-1 bodyMd-medium text-text-default">
              Total tags
            </div>
            <div className="text-text-soft bodyMd">
              {values.tags.length || 0}
            </div>
          </div>
        </div>
      </ReviewComponent>
      {values.advanceOptions &&
        (Object.keys(values.buildArgs).length > 0 ||
          Object.keys(values.buildContexts).length > 0 ||
          values.contextDir ||
          values.dockerfilePath) && (
          <ReviewComponent title="Advance options" onEdit={() => onEdit(2)}>
            <div className="flex flex-col gap-xl p-xl rounded border border-border-default">
              {Object.keys(values.buildArgs).length > 0 && (
                <div className="flex flex-row items-center gap-lg [&:not(:last-child)]:pb-lg [&:not(:last-child)]:border-b border-border-default">
                  <div className="flex-1 bodyMd-medium text-text-default">
                    Build Args
                  </div>
                  <div className="text-text-soft bodyMd">
                    {Object.keys(values.buildArgs).length || 0}
                  </div>
                </div>
              )}
              {Object.keys(values.buildContexts).length > 0 && (
                <div className="flex flex-row items-center gap-lg [&:not(:last-child)]:pb-lg [&:not(:last-child)]:border-b border-border-default">
                  <div className="flex-1 bodyMd-medium text-text-default">
                    Build contexts
                  </div>
                  <div className="text-text-soft bodyMd">
                    {Object.keys(values.buildContexts).length || 0}
                  </div>
                </div>
              )}
              {values.contextDir && (
                <div className="flex flex-col gap-md [&:not(:last-child)]:pb-lg [&:not(:last-child)]:border-b border-border-default">
                  <div className="bodyMd-medium text-text-default">
                    Context dir
                  </div>
                  <div className="bodySm text-text-soft">
                    {values.contextDir}
                  </div>
                </div>
              )}
              {values.dockerfilePath && (
                <div className="flex flex-col gap-md [&:not(:last-child)]:pb-lg [&:not(:last-child)]:border-b border-border-default">
                  <div className="bodyMd-medium text-text-default">
                    Docker filepath
                  </div>
                  <div className="bodySm text-text-soft">
                    {values.dockerfilePath}
                  </div>
                </div>
              )}
            </div>
          </ReviewComponent>
        )}
    </div>
  );
};

export default ReviewBuild;
