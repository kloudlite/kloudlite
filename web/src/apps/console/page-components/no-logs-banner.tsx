import { motion } from 'framer-motion';
import Pulsable from '@kloudlite/design-system/atoms/pulsable';
import { EmptyState } from '~/console/components/empty-state';
import { SmileySad } from '~/console/components/icons';
import Wrapper from '~/console/components/wrapper';

export const NoLogsAndMetricsBanner = ({
  title,
  description,
}: {
  title: string;
  description: string;
}) => {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ ease: 'anticipate', duration: 0.1 }}
      className="flex flex-col py-4xl"
    >
      <Wrapper>
        <Pulsable isLoading={false}>
          <EmptyState
            {...{
              image: <SmileySad size={48} />,
              heading: <span>{title}</span>,
              footer: (
                <span className="flex items-center justify-center text-sm">
                  {description}
                </span>
              ),
            }}
          />
        </Pulsable>
      </Wrapper>
    </motion.div>
  );
};
