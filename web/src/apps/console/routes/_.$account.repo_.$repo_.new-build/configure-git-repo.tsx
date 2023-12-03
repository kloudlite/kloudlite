import { useState } from 'react';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { AnimatePresence, motion } from 'framer-motion';
import { FadeIn } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/util';

const ConfigureRepo = () => {
  const [setting, setSetting] = useState('general');
  return (
    <FadeIn notForm>
      <div className="flex flex-col gap-xl">
        <div className="headingXl text-text-default">
          Configure git repository
        </div>
        <ExtendedFilledTab
          size="sm"
          items={[
            {
              value: 'general',
              label: 'General',
            },

            {
              value: 'advance',
              label: 'Advance settings',
            },
          ]}
          value={setting}
          onChange={(e) => setSetting(e)}
        />
      </div>
      <AnimatePresence mode="wait">
        <motion.div
          key={setting || 'empty'}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="flex flex-col gap-6xl w-full"
        >
          {setting === 'general' && <div>general</div>}
          {setting === 'advance' && <div>advance</div>}
        </motion.div>
      </AnimatePresence>
    </FadeIn>
  );
};

export default ConfigureRepo;
