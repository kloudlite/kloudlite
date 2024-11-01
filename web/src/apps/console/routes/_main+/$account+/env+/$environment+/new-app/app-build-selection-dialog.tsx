import { Dispatch, SetStateAction, useState } from 'react';
import { IconButton } from '@kloudlite/design-system/atoms/button';
import Popup from '@kloudlite/design-system/molecule/popup';
import { toast } from '@kloudlite/design-system/molecule/toast';
import { titleCase, useMapper } from '@kloudlite/design-system/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import { ArrowLeft, Spinner } from '~/console/components/icons';
import ListV2 from '~/console/components/listV2';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IBuilds } from '~/console/server/gql/queries/build-queries';
import { IRepos } from '~/console/server/gql/queries/repo-queries';
import {
  ExtractNodeType,
  parseNodes,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';

type IBuildType = ExtractNodeType<IBuilds>;
type IRepoType = ExtractNodeType<IRepos>;

type IBuildSelectionDialog = {
  label: string;
  value: string;
  build?: IBuildType;
  repo?: IRepoType;
  updateInfo: { author: string; time: string } | null;
};

const BuildSelectDialogList = ({
  data,
  onClick,
  type,
  selectedOption,
}: {
  data: IBuildSelectionDialog[];
  onClick: (data: IBuildSelectionDialog) => void;
  type: 'repo' | 'builds';
  selectedOption: string;
}) => {
  return (
    <ListV2.Root
      data={{
        headers: [
          {
            name: 'name',
            render: () => 'Name',
            className: 'flex-1',
          },

          {
            render: () => (type === 'repo' ? 'Updated' : null),
            name: 'updated',
            className: 'w-[180px]',
          },
        ],
        rows: data.map((r) => {
          return {
            onClick: () => onClick(r),
            pressed: r.value === selectedOption,
            columns: {
              name: {
                render: () => <ListTitle title={r.label} />,
              },
              ...(type === 'repo'
                ? {
                    updated: {
                      render: () => (
                        <ListItem
                          data={`${r.updateInfo?.author}`}
                          subtitle={r.updateInfo?.time}
                        />
                      ),
                    },
                  }
                : {}),
            },
          };
        }),
      }}
    />
  );
};

const BuildSelectionDialog = ({
  open,
  setOpen,
  onChange,
}: {
  open: boolean;
  setOpen: Dispatch<SetStateAction<boolean>>;
  onChange?: (data: { build?: ExtractNodeType<IBuilds> }) => void;
}) => {
  const [repoName, setRepoName] = useState('');
  const [build, setBuild] = useState<ExtractNodeType<IBuilds> | null>();

  const {
    currentStep,
    onNext,
    onPrevious,
    reset: resetStep,
  } = useMultiStep({
    totalSteps: 2,
    defaultStep: 1,
  });

  const api = useConsoleApi();

  const {
    data,
    isLoading: repoLoading,
    error: repoLoadingError,
  } = useCustomSwr('/repos', async () => {
    return api.listRepo({});
  });

  const repos = useMapper(parseNodes(data), (val) => ({
    label: val.name,
    value: val.name,
    repo: val,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(val))}`,
      time: parseUpdateOrCreatedOn(val),
    },
  }));

  const {
    data: buildData,
    isLoading: buildLoading,
    error: buildError,
  } = useCustomSwr(
    () => (repoName ? `/build_${repoName}` : null),
    async () => {
      return api.listBuilds({ repoName });
    }
  );

  const builds = useMapper(parseNodes(buildData), (val) => ({
    label: val.name,
    value: val.id,
    build: val,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(val))}`,
      time: parseUpdateOrCreatedOn(val),
    },
  }));

  const reset = () => {
    setRepoName('');
    setBuild(null);
    resetStep();
  };

  return (
    <Popup.Root show={open} onOpenChange={setOpen} className="!w-[900px]">
      <Popup.Header showclose={false}>
        <div className="flex flex-row items-center gap-lg">
          {currentStep === 2 && (
            <IconButton
              size="sm"
              icon={<ArrowLeft />}
              variant="plain"
              onClick={() => {
                onPrevious();
                reset();
              }}
            />
          )}
          <div className="flex-1">
            {currentStep === 1 ? 'Select Repository' : 'Select build'}
          </div>
          <div className="bodyMd text-text-strong font-normal">
            {currentStep}/2
          </div>
        </div>
      </Popup.Header>
      <Popup.Content>
        <div className="min-h-[40vh]">
          {repoLoading || buildLoading ? (
            <div className="flex flex-col items-center justify-center gap-xl pt-5xl">
              <span className="animate-spin">
                <Spinner color="currentColor" weight={2} size={24} />
              </span>
              <span className="text-text-soft bodyMd">Loading</span>
            </div>
          ) : (
            <MultiStep.Root currentStep={currentStep}>
              <MultiStep.Step step={1}>
                <BuildSelectDialogList
                  selectedOption=""
                  data={repos}
                  onClick={(d) => {
                    setRepoName(d.value);
                    onNext();
                  }}
                  type="repo"
                />
              </MultiStep.Step>
              <MultiStep.Step step={2}>
                <BuildSelectDialogList
                  data={builds}
                  selectedOption={build?.id || ''}
                  onClick={(t) => {
                    setBuild(t.build);
                  }}
                  type="builds"
                />
              </MultiStep.Step>
            </MultiStep.Root>
          )}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button
          closable={currentStep === 1}
          content={currentStep === 1 ? 'Cancel' : 'Back'}
          variant="basic"
          onClick={() => {
            onPrevious();
            reset();
          }}
        />
        <Popup.Button
          type="submit"
          disabled={currentStep === 1 || (!build && currentStep === 2)}
          content={currentStep === 2 ? 'Select' : 'Continue'}
          variant="primary"
          onClick={() => {
            if (build) {
              onChange?.({ build });
              reset();
              setOpen(false);
            } else {
              toast.error('Something went wrong.');
            }
          }}
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

export default BuildSelectionDialog;
