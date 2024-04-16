import { AnimatePresence, motion } from 'framer-motion';
import { ReactNode, useState } from 'react';
import ActionList from '~/components/atoms/action-list';

interface IItem {
  label: string;
  value: string;
  prefix?: JSX.Element;
  panel: ReactNode;
}
interface IPopupSidebarLayout {
  items: Array<IItem>;
  defaultActive?: IItem;
}
const PopupSidebarLayout = ({
  items = [],
  defaultActive,
}: IPopupSidebarLayout) => {
  const [activePanel, setActivePanel] = useState<IItem | undefined>(
    defaultActive || items[0]
  );
  return (
    <div className="flex flex-row items-start min-h-[50vh]">
      <div className="pr-3xl min-w-[180px]">
        <ActionList.Root
          value={activePanel?.value || ''}
          showIndicator={false}
          onChange={(value) =>
            setActivePanel(items.find((i) => i.value === value))
          }
          className="!gap-lg"
        >
          {items.map((ai) => (
            <ActionList.Item prefix={ai.prefix} key={ai.value} value={ai.value}>
              {ai.label}
            </ActionList.Item>
          ))}
        </ActionList.Root>
      </div>
      <AnimatePresence mode="wait">
        <motion.div
          key={activePanel?.value}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ ease: 'anticipate', duration: 0.2 }}
          className="flex-1"
        >
          <div className="flex-1">{activePanel?.panel}</div>
        </motion.div>
      </AnimatePresence>
    </div>
  );
};

export default PopupSidebarLayout;
