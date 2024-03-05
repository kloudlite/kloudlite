import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { AnimatePresence, motion } from 'framer-motion';
import {
  createAppEnvPage,
  useAppState,
} from '~/console/page-components/app-states';
import { BottomNavigation } from '~/console/components/commons';
import { FadeIn } from '../../../../../../page-components/util';
import { EnvironmentVariables } from './app-environment-variables';
import { ConfigMounts } from './app-environment-mounts';

export interface IAppDialogValue {
  refKey: string;
  refName: string;
  type: 'config' | 'secret';
}

const AppEnvironment = () => {
  const { envPage, setEnvPage, setPage, markPageAsCompleted } = useAppState();
  const items: {
    label: string;
    value: createAppEnvPage;
  }[] = [
    {
      label: `Environment variables`,
      value: 'environment_variables',
    },
    {
      label: 'Config mount',
      value: 'config_mounts',
    },
  ];

  return (
    <FadeIn notForm>
      <div className="flex flex-col gap-xl ">
        <ExtendedFilledTab
          value={envPage}
          onChange={setEnvPage}
          items={items}
        />
      </div>
      <AnimatePresence mode="wait">
        <motion.div
          key={envPage || 'empty'}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="flex flex-col gap-6xl w-full"
        >
          {envPage === 'environment_variables' && <EnvironmentVariables />}
          {envPage === 'config_mounts' && <ConfigMounts />}
        </motion.div>
      </AnimatePresence>

      <BottomNavigation
        primaryButton={{
          type: 'submit',
          content: 'Save & Continue',
          variant: 'primary',
          onClick: () => {
            setPage(4);
            markPageAsCompleted(3);
          },
        }}
        secondaryButton={{
          content: 'Compute',
          variant: 'outline',
          onClick: () => {
            setPage(2);
          },
        }}
      />
    </FadeIn>
  );
};

export default AppEnvironment;
