import { useParams } from '@remix-run/react';
import { useCallback, useEffect, useState } from 'react';
import { Checkbox } from '~/components/atoms/checkbox';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { BottomNavigation } from '~/console/components/commons';
import { NameIdView } from '~/console/components/name-id-view';
import { useAppState } from '~/console/page-components/app-states';
import { FadeIn } from '~/console/page-components/util';
import { AppSelectItem } from '~/console/routes/_main+/$account+/env+/$environment+/new-app/app-detail';
import HandleBuild from '~/console/routes/_main+/$account+/repo+/$repo+/builds/handle-builds';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import { constants } from '~/console/server/utils/constants';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import {
  DISCARD_ACTIONS,
  useUnsavedChanges,
} from '~/root/lib/client/hooks/use-unsaved-changes';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

// const ExtraButton = ({
//   onNew,
//   onEdit,
//   onTrigger,
//   isExistingBuild,
// }: {
//   onNew: () => void;
//   onEdit: () => void;
//   onTrigger: () => void;
//   isExistingBuild: boolean;
// }) => {
//   let options: IResourceExtraItem[] = [];

//   if (isExistingBuild) {
//     options = [
//       {
//         label: 'Connect new',
//         icon: <PencilSimple size={16} />,
//         type: 'item',
//         key: 'new',
//         onClick: onNew,
//       },
//       ...options,
//     ];
//   } else {
//     options = [
//       {
//         label: 'Edit',
//         icon: <PencilSimple size={16} />,
//         type: 'item',
//         key: 'edit',
//         onClick: onEdit,
//       },
//       {
//         label: 'Trigger',
//         icon: <ArrowClockwise size={16} />,
//         type: 'item',
//         key: 'trigger',
//         onClick: onTrigger,
//       },
//       ...options,
//     ];
//   }
//   return <ResourceExtraAction options={options} />;
// };

const valueRenderer = ({ value }: { value: string }) => {
  return <div>{value}</div>;
};

const AppGeneral = ({ mode = 'new' }: { mode: 'edit' | 'new' }) => {
  const {
    readOnlyApp,
    setApp,
    setPage,
    markPageAsCompleted,
    activeContIndex,
    setContainer,
  } = useAppState();

  const { performAction } = useUnsavedChanges();

  // const [openBuildSelection, setOpenBuildSelection] = useState(false);

  // only for edit mode
  // const [isEdited, setIsEdited] = useState(!app.ciBuildId);
  const [showBuildEdit, setShowBuildEdit] = useState(false);
  const api = useConsoleApi();
  const params = useParams();

  const [imageList, setImageList] = useState<any[]>([]);
  const [imageLoaded, setImageLoaded] = useState(false);
  const [imageSearchText, setImageSearchText] = useState('');

  const getRegistryImages = useCallback(
    async ({ query }: { query: string }) => {
      ensureAccountClientSide(params);
      setImageLoaded(true);
      try {
        const registrayImages = await api.searchRegistryImages({ query });
        const data = registrayImages.data.map((i) => ({
          label: `${i.imageName}:${i.imageTag}/${i.meta.author}/${i.meta.registry}:${i.meta.repository}`,
          value: `${i.imageName}:${i.imageTag}`,
          ready: true,
          render: () => (
            <AppSelectItem
              label={`${i.imageName}:${i.imageTag}`}
              meta={i.meta}
              imageSearchText={query}
            />
          ),
        }));
        setImageList(data);
      } catch (err) {
        handleError(err);
      } finally {
        setImageLoaded(false);
      }
    },
    [],
  );

  useEffect(() => {
    getRegistryImages({ query: '' });
  }, []);

  useDebounce(
    () => {
      if (imageSearchText) {
        getRegistryImages({ query: imageSearchText });
      }
    },
    300,
    [imageSearchText],
  );

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    isLoading,
    submit,
    resetValues,
  } = useForm({
    initialValues: {
      name: parseName(readOnlyApp),
      displayName: readOnlyApp.displayName,
      isNameError: false,
      imageMode:
        readOnlyApp.metadata?.annotations[keyconstants.appImageMode] ||
        'default',
      imageUrl: readOnlyApp?.spec.containers[activeContIndex]?.image || '',
      manualRepo: '',
      source: {
        branch: readOnlyApp?.build?.source.branch,
        repository: readOnlyApp?.build?.source.repository,
        provider: readOnlyApp?.build?.source.provider,
      },
      advanceOptions: false,
      buildArgs: {},
      buildContexts: {},
      contextDir: '',
      dockerfilePath: '',
      dockerfileContent: '',
      isGitLoading: false,

      imagePullPolicy:
        readOnlyApp?.spec.containers[activeContIndex]?.imagePullPolicy ||
        'Always',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      imageUrl: Yup.string().matches(
        constants.dockerImageFormatRegex,
        'Invalid image format',
      ),
      manualRepo: Yup.string().when(
        ['imageUrl', 'imageMode'],
        ([imageUrl, imageMode], schema) => {
          const regex = /^(\w+):(\w+)$/;
          if (imageMode === 'git') {
            return schema;
          }
          if (!imageUrl) {
            return schema.required().matches(regex, 'Invalid image format');
          }
          return schema;
        },
      ),
      imageMode: Yup.string().required(),
      source: Yup.object()
        .shape({})
        .test('is-empty', 'Branch is required.', (v, c) => {
          // @ts-ignoredfgdfg
          if (!v?.branch && c.parent.imageMode === 'git') {
            return false;
          }
          return true;
        }),
    }),

    onSubmit: async (val) => {
      setApp((a) => {
        return {
          ...a,
          metadata: {
            ...a.metadata,
            annotations: {
              ...(a.metadata?.annotations || {}),
              [keyconstants.appImageMode]: val.imageMode,
            },
            name: val.name,
          },
          displayName: val.displayName,
          spec: {
            ...a.spec,
            containers: [
              {
                ...a.spec.containers?.[0],
                image: val.imageUrl || val.manualRepo,
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

  /** ---- Only for edit mode in settings ----* */
  useEffect(() => {
    if (mode === 'edit') {
      submit();
    }
  }, [values, mode]);

  useEffect(() => {
    if (performAction === DISCARD_ACTIONS.DISCARD_CHANGES) {
      // if (app.ciBuildId) {
      //   setIsEdited(false);
      // }
      resetValues();
      // @ts-ignore
      // setBuildData(readOnlyApp?.build);
    }

    // else if (performAction === 'init') {
    //   setIsEdited(false);
    // }
  }, [performAction]);

  useEffect(() => {
    resetValues();
  }, [readOnlyApp]);

  useEffect(() => {
    console.log('values', values.imagePullPolicy);
    // if (
    //   values.imagePullPolicy !==
    //   readOnlyApp.spec.containers[activeContIndex].imagePullPolicy
    // ) {
    setContainer((s) => {
      return {
        ...s,
        imagePullPolicy: values.imagePullPolicy,
      };
    });
    // }
  }, [values.imagePullPolicy]);

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
      {mode === 'new' && (
        <div className="bodyMd text-text-soft">
          The application streamlines project management through intuitive task
          tracking and collaboration tools.
        </div>
      )}
      <div className="flex flex-col gap-5xl">
        {mode === 'new' ? (
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
        ) : (
          <TextInput
            value={values.displayName}
            label="Name"
            size="lg"
            onChange={handleChange('displayName')}
          />
        )}
        <div className="flex flex-col gap-xl">
          {/* <TextInput
            size="lg"
            label="Image name"
            placeholder="Enter Image name"
            value={values.imageUrl}
            onChange={handleChange('imageUrl')}
            error={!!errors.imageUrl}
            message={errors.imageUrl}
          /> */}

          <Select
            label="Select Images"
            size="lg"
            value={values.imageUrl}
            placeholder="Select a image"
            creatable
            searchable
            options={async () => imageList}
            onChange={({ value }) => {
              handleChange('imageUrl')(dummyEvent(value));
            }}
            onSearch={(text) => {
              setImageSearchText(text);
            }}
            showclear
            noOptionMessage={
              <div className="p-2xl bodyMd text-center">
                Search for image or enter image name
              </div>
            }
            error={!!errors.imageUrl}
            message={errors.imageUrl}
            loading={imageLoaded}
            createLabel="Select"
            valueRender={valueRenderer}
          />
        </div>

        <Checkbox
          label="Always pull image on restart"
          checked={values.imagePullPolicy === 'Always'}
          onChange={(val) => {
            const imagePullPolicy = val ? 'Always' : 'IfNotPresent';
            handleChange('imagePullPolicy')(dummyEvent(imagePullPolicy));
          }}
        />
      </div>
      {mode === 'new' && (
        <BottomNavigation
          primaryButton={{
            loading: isLoading,
            type: 'submit',
            content: 'Save & Continue',
            variant: 'primary',
          }}
        />
      )}
      <HandleBuild
        {...{
          isUpdate: true,
          data: { mode: 'app', ...readOnlyApp.build! },
          visible: showBuildEdit,
          setVisible: () => setShowBuildEdit(false),
        }}
      />
    </FadeIn>
  );
};

export default AppGeneral;
