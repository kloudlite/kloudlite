import { ArrowRight, ConfigurationFill, UserCircle } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { useMapper } from '~/components/utils';
import RawWrapper, { TitleBox } from '~/console/components/raw-wrapper';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import AlertModal from '~/console/components/alert-modal';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { FadeIn } from '~/console/page-components/util';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/console/components/multi-step-progress';
import MultiStepProgressWrapper from '~/console/components/multi-step-progress-wrapper';
import BuildDetails from './build-details';
import RepoSelector from './repo-selector';
import ConfigureRepo from './configure-git-repo';

const NewBuild = () => {
  const navigate = useNavigate();

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);

  const { a: accountName } = useParams();

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      description: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      description: Yup.string().required(),
    }),
    onSubmit: async () => {
      try {
        console.log();
      } catch (err) {
        handleError(err);
      }
    },
  });

  const items = useMapper(
    [
      {
        label: 'Build details',
        active: true,
        id: 1,
        completed: false,
      },
      {
        label: 'Import Git Repository',
        active: false,
        id: 2,
        completed: false,
      },
      {
        label: 'Configure build',
        active: false,
        id: 3,
        completed: false,
      },
    ],
    (i) => {
      return {
        value: i.id,
        item: {
          ...i,
        },
      };
    }
  );

  const { currentStep, nextStep, prevStep, reset, jumpStep } =
    useMultiStepProgress({
      defaultStep: 1,
      totalSteps: 3,
    });

  return (
    // <>
    //   <RawWrapper
    //     title="Let’s create new build."
    //     subtitle="Create your build under to repo effortlessly"
    //     progressItems={items}
    //     badge={{
    //       title: 'Kloudlite Labs Pvt Ltd',
    //       subtitle: accountName,
    //       image: <UserCircle size={20} />,
    //     }}
    //     onCancel={() => setShowUnsavedChanges(true)}
    //     rightChildren={
    //       <FadeIn onSubmit={handleSubmit}>
    //         <MultiStep.Root currentStep={currentStep}>
    //           <MultiStep.Step
    //             step={1}
    //             className="flex flex-col gap-6xl w-full justify-center"
    //           >
    //             <BuildDetails
    //               errors={errors}
    //               values={values}
    //               handleChange={handleChange}
    //             />
    //           </MultiStep.Step>
    //           <MultiStep.Step
    //             step={2}
    //             className="flex flex-col gap-6xl w-full justify-center"
    //           >
    //             <RepoSelector />
    //           </MultiStep.Step>
    //           <MultiStep.Step
    //             step={3}
    //             className="flex flex-col gap-6xl w-full justify-center"
    //           >
    //             <ConfigureRepo />
    //           </MultiStep.Step>
    //         </MultiStep.Root>
    //         <div className="flex flex-row gap-xl justify-end items-center">
    //           <Button
    //             loading={isLoading}
    //             content="Continue"
    //             suffix={<ArrowRight />}
    //             variant="primary"
    //             onClick={onNext}
    //           />
    //         </div>
    //       </FadeIn>
    //     }
    //   />

    //   <AlertModal
    //     title="Leave page with unsaved changes?"
    //     message="Leaving this page will delete all unsaved changes."
    //     okText="Leave page"
    //     cancelText="Stay"
    //     variant="critical"
    //     show={showUnsavedChanges}
    //     setShow={setShowUnsavedChanges}
    //     onSubmit={() => {
    //       setShowUnsavedChanges(false);
    //       navigate(`/${accountName}/projects`);
    //     }}
    //   />
    // </>
    <MultiStepProgressWrapper
      title="Create new Build"
      subTitle="Create your build under to repo effortlessly"
    >
      <MultiStepProgress.Root currentStep={currentStep} jumpStep={jumpStep}>
        <MultiStepProgress.Step label="Build details" step={1}>
          <div className="flex flex-col gap-3xl">
            <BuildDetails
              errors={errors}
              values={values}
              handleChange={handleChange}
            />
            <Button
              content="Next"
              variant="primary"
              onClick={nextStep}
              size="lg"
              suffix={<ArrowRight />}
            />
          </div>
        </MultiStepProgress.Step>
        <MultiStepProgress.Step label="Import git repository" step={2}>
          <RepoSelector />
        </MultiStepProgress.Step>
      </MultiStepProgress.Root>
    </MultiStepProgressWrapper>
  );
};

export default NewBuild;

export const handle = {
  noMainLayout: true,
};
