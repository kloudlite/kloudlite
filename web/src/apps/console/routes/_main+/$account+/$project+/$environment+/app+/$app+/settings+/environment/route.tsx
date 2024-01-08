import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { AnimatePresence, motion } from 'framer-motion';
import {
  createAppEnvPage,
  useAppState,
} from '~/console/page-components/app-states';
import Wrapper from '~/console/components/wrapper';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { Button } from '~/components/atoms/button';
import { EnvironmentVariables } from '../../../../new-app/app-environment-variables';
import { ConfigMounts } from '../../../../new-app/app-environment-mounts';

export interface IAppDialogValue {
  refKey: string;
  refName: string;
  type: 'config' | 'secret';
}

const SettingEnvironment = () => {
  const { envPage, setEnvPage } = useAppState();
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();
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
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'Environment',
          action: hasChanges && (
            <div className="flex flex-row items-center gap-lg">
              <Button
                disabled={loading}
                variant="basic"
                content="Discard changes"
                onClick={() => setPerformAction('discard-changes')}
              />
              <Button
                disabled={loading}
                content={loading ? 'Committing changes.' : 'View changes'}
                loading={loading}
                onClick={() => setPerformAction('view-changes')}
              />
            </div>
          ),
        }}
      >
        <ExtendedFilledTab
          value={envPage}
          onChange={setEnvPage}
          items={items}
        />
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
      </Wrapper>
    </div>
  );
};

export default SettingEnvironment;
