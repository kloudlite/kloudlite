import { ChangeEventHandler, Dispatch, SetStateAction, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import { Chip, ChipGroup } from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import Tooltip from '~/components/atoms/tooltip';
import Popup from '~/components/molecule/popup';
import { titleCase, useMapper } from '~/components/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import { ArrowLeft, Spinner, XCircleFill } from '~/console/components/icons';
import ListV2 from '~/console/components/listV2';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import {
  parseNodes,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';

type IRepoSelectDialogListItem = {
  label: string;
  value: string;
  updateInfo: { author: string; time: string } | null;
};
const RepoSelectDialogList = ({
  data,
  onClick,
  type,
  selectedOption,
}: {
  data: IRepoSelectDialogListItem[];
  onClick: (data: IRepoSelectDialogListItem) => void;
  type: 'repo' | 'tags';
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

const RepoSelectDialog = ({
  open,
  setOpen,
  onChange,
}: {
  open: boolean;
  setOpen: Dispatch<SetStateAction<boolean>>;
  onChange?: (data: { repo: string; tag: string }) => void;
}) => {
  const [repoName, setRepoName] = useState('');
  const [tag, setTag] = useState('');

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
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(val))}`,
      time: parseUpdateOrCreatedOn(val),
    },
  }));

  const {
    data: digestData,
    isLoading: digestLoading,
    error: digestError,
  } = useCustomSwr(
    () => (repoName ? `/digests_${repoName}` : null),
    async () => {
      return api.listDigest({ repoName });
    }
  );

  const tags = useMapper(parseNodes(digestData), (val) => val.tags)
    .flat()
    .flatMap((f) => ({ label: f, value: f, updateInfo: null }));

  const reset = () => {
    setRepoName('');
    setTag('');
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
            {currentStep === 1 ? 'Select Repository' : 'Select tag'}
          </div>
          <div className="bodyMd text-text-strong font-normal">
            {currentStep}/2
          </div>
        </div>
      </Popup.Header>
      <Popup.Content>
        <div className="min-h-[40vh]">
          {repoLoading || digestLoading ? (
            <div className="flex flex-col items-center justify-center gap-xl pt-5xl">
              <span className="animate-spin">
                <Spinner color="currentColor" weight={2} size={24} />
              </span>
              <span className="text-text-soft bodyMd">Loading</span>
            </div>
          ) : (
            <MultiStep.Root currentStep={currentStep}>
              <MultiStep.Step step={1}>
                <RepoSelectDialogList
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
                <RepoSelectDialogList
                  data={tags}
                  selectedOption={tag}
                  onClick={(t) => {
                    setTag(t.value);
                  }}
                  type="tags"
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
          disabled={currentStep === 1 || (!tag && currentStep === 2)}
          content={currentStep === 2 ? 'Select' : 'Continue'}
          variant="primary"
          onClick={() => {
            onChange?.({ repo: repoName, tag });
            reset();
            setOpen(false);
          }}
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

const RepoSelector = ({
  tag,
  repo,
  onClear,
  textValue,
  onTextChanged,
  onValueChange,
  error,
}: {
  tag: string;
  repo: string;
  onClear: () => void;
  onTextChanged: ChangeEventHandler<HTMLInputElement> | undefined;
  textValue: string;
  onValueChange: (data: { repo: string; tag: string }) => void;
  error?: string;
}) => {
  const [open, setOpen] = useState(false);
  return (
    <div className="">
      {tag && repo && !textValue ? (
        <div className="flex flex-col max-w-full">
          <div className="bodyMd-medium text-text-default pb-md">
            <span className="h-4xl block flex items-center">Image</span>
          </div>
          <div className="max-h-[43px] px-lg flex flex-row items-center rounded border border-border-default bg-surface-basic-default line-clamp-1 justify-between">
            <Tooltip.Root
              className="!max-w-[400px]"
              content={
                <div className="flex-1 flex flex-row gap-md items-center py-xl px-lg bodyMd text-text-soft ">
                  <span className="line-clamp-1">{repo}</span>
                  <span>:</span>

                  <span className="line-clamp-1">{tag}</span>
                </div>
              }
            >
              <div className="gap-md items-center py-xl px-md bodyMd text-text-soft max-w-full w-fit">
                <div className="flex flex-row items-center truncate">
                  <span className="truncate">{repo}</span>
                  <span>:</span>
                  <span className="truncate">{tag}</span>
                </div>
              </div>
            </Tooltip.Root>
            <button
              aria-label="clear"
              tabIndex={-1}
              type="button"
              className="outline-none p-lg text-text-default rounded-full"
              onClick={onClear}
            >
              <XCircleFill size={16} color="currentColor" />
            </button>
          </div>
        </div>
      ) : (
        <TextInput
          placeholder="eg: nginx:latest"
          value={textValue}
          onChange={onTextChanged}
          label="Image"
          size="lg"
          error={!!error}
          message={error}
          suffix={
            !textValue ? (
              <ChipGroup onClick={() => setOpen(true)}>
                <Chip
                  label="Kloudlite repository"
                  item="repo"
                  type="CLICKABLE"
                />
              </ChipGroup>
            ) : null
          }
          showclear={!!textValue}
        />
      )}
      <RepoSelectDialog
        open={open}
        setOpen={setOpen}
        onChange={onValueChange}
      />
    </div>
  );
};

export default RepoSelector;
