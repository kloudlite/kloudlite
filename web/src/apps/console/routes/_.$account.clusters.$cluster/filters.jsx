import { AnimatePresence, motion } from 'framer-motion';
import { cn } from '~/components/utils';
import ScrollArea from '~/console/components/scroll-area';
import * as Chips from '~/components/atoms/chips';
import { ChipGroupPaddingTop } from '~/design-system/tailwind-base';
import { Button } from '~/components/atoms/button';

const Filters = ({ appliedFilters, setAppliedFilters }) => {
  return (
    <AnimatePresence initial={false}>
      {appliedFilters.length > 0 && (
        <motion.div
          className={cn('flex flex-row gap-xl relative')}
          initial={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          animate={{
            height: '46px',
            opacity: 1,
            paddingTop: ChipGroupPaddingTop,
          }}
          exit={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          transition={{
            ease: 'linear',
          }}
          onAnimationStart={(e) => console.log(e)}
        >
          <ScrollArea className="flex-1">
            <Chips.ChipGroup
              onRemove={(c) =>
                setAppliedFilters(appliedFilters.filter((a) => a.id !== c))
              }
            >
              {appliedFilters.map((af) => {
                return <Chips.Chip {...af} key={af.id} item={af} />;
              })}
            </Chips.ChipGroup>
          </ScrollArea>
          {appliedFilters.length > 0 && (
            <div className="flex flex-row items-center justify-center">
              <Button
                content="Clear all"
                variant="primary-plain"
                onClick={() => {
                  setAppliedFilters([]);
                }}
              />
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default Filters;
