import { TextInput } from '~/components/atoms/input';
import { useAppState } from '~/console/page-components/app-states';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import {
  parseName,
  parseNodes,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { FadeIn } from '~/console/page-components/util';
import { NameIdView } from '~/console/components/name-id-view';
import { BottomNavigation } from '~/console/components/commons';
import { registryHost } from '~/lib/configs/base-url.cjs';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { titleCase, useMapper } from '~/components/utils';
import Tooltip from '~/components/atoms/tooltip';
import { Chip, ChipGroup } from '~/components/atoms/chips';
import { ArrowLeft, Spinner, XCircleFill } from '~/console/components/icons';
import Popup from '~/components/molecule/popup';
import { Dispatch, SetStateAction, useState } from 'react';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { IconButton } from '~/components/atoms/button';
import ListV2 from '~/console/components/listV2';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import { useParams } from '@remix-run/react';

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
        <div className="min-h-[40vh] h-full">
          {true ? (
            <div className="flex flex-col items-center justify-center gap-xl h-full">
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
}: {
  tag: string;
  repo: string;
  onClear: () => void;
  onTextChanged: () => void;
  textValue: string;
  onValueChange: (data: { repo: string; tag: string }) => void;
}) => {
  const [open, setOpen] = useState(false);
  return (
    <div className="flex-1">
      {tag && repo && !textValue ? (
        <div className="flex flex-col">
          <div className="bodyMd-medium text-text-default pb-md">
            <span className="h-4xl block flex items-center">Repository</span>
          </div>
          <div className="max-h-[43px] px-lg flex flex-row items-center rounded border border-border-default bg-surface-basic-default line-clamp-1">
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
              <div className="flex-1 flex flex-row gap-md items-center py-xl px-md bodyMd text-text-soft ">
                <span className="line-clamp-1">{repo}</span>
                <span>:</span>

                <span className="line-clamp-1">{tag}</span>
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
          label="Repository"
          size="lg"
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
const AppDetail = () => {
  const { app, setApp, setPage, markPageAsCompleted, activeContIndex } =
    useAppState();

  const { account } = useParams();

  const { values, errors, handleChange, handleSubmit, isLoading, setValues } =
    useForm({
      initialValues: {
        name: parseName(app),
        displayName: app.displayName,
        isNameError: false,
        imageUrl: app.spec.containers[activeContIndex]?.image || '',
        manualRepo: '',
        repoAccountName:
          app.metadata?.annotations?.[keyconstants.repoAccountName] || '',
      },
      validationSchema: Yup.object({
        name: Yup.string().required(),
        displayName: Yup.string().required(),
        repoName: Yup.string().required(),
        repoImageTag: Yup.string().required(),
      }),

      onSubmit: async (val) => {
        setApp((a) => {
          return {
            ...a,
            metadata: {
              ...a.metadata,
              annotations: {
                ...(a.metadata?.annotations || {}),
                [keyconstants.repoAccountName]: val.repoAccountName,
              },
              name: val.name,
            },
            displayName: val.displayName,
            spec: {
              containers: [
                {
                  ...(a.spec.containers?.[0] || {}),
                  image: val.imageUrl,
                  name: 'container-0',
                },
              ],
            },
          };
        });
        setPage(2);
        markPageAsCompleted(1);
      },
    });

  return (
    <FadeIn
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <div className="bodyMd text-text-soft">
        The application streamlines project management through intuitive task
        tracking and collaboration tools.
      </div>
      <div className="flex flex-col gap-3xl">
        <NameIdView
          displayName={values.displayName}
          name={values.name}
          resType="app"
          errors={errors.name}
          label="Application name"
          placeholder="Enter application name"
          handleChange={handleChange}
          nameErrorLabel="isNameError"
        />
        <RepoSelector
          tag={values.imageUrl.split(':')[1]}
          repo={
            values.imageUrl
              .replace(`${registryHost}/${account}/`, '')
              .split(':')[0]
          }
          onClear={() => {
            setValues((v) => {
              return {
                ...v,
                imageUrl: '',
                manualRepo: '',
              };
            });
          }}
          textValue={values.manualRepo}
          onTextChanged={handleChange('manualRepo')}
          onValueChange={({ repo, tag }) => {
            handleChange('imageUrl')(
              dummyEvent(`${registryHost}/${account}/${repo}:${tag}`)
            );
          }}
        />
      </div>
      <BottomNavigation
        primaryButton={{
          loading: isLoading,
          type: 'submit',
          content: 'Save & Continue',
          variant: 'primary',
        }}
      />
    </FadeIn>
  );
};

export default AppDetail;
