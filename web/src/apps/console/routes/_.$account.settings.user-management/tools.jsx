import { ArrowDown, ArrowUp, ArrowsDownUp } from '@jengaicons/react';
import { useState } from 'react';
import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';

const Tools = () => {
  return (
    <div className="flex-1 ">
      <Toolbar.Root>
        <div className="w-full flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <Toolbar.TextInput placeholder="Search" value="" />
          </div>
          <SortbyOptionList />
        </div>
      </Toolbar.Root>
    </div>
  );
};
const SortbyOptionList = () => {
  const [sortbyProperty, setSortbyProperty] = useState('updated');
  const [sortbyTime, setSortbyTime] = useState('oldest');
  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            <Toolbar.IconButton icon={<ArrowsDownUp />} variant="basic" />
          </div>

          <div className="flex md:hidden">
            <Toolbar.IconButton variant="basic" icon={<ArrowsDownUp />} />
          </div>
        </div>
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.RadioGroup
          value={sortbyProperty}
          onValueChange={setSortbyProperty}
        >
          <OptionList.RadioGroupItem
            value="title"
            onSelect={(e) => e.preventDefault()}
          >
            Name
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="updated"
            onSelect={(e) => e.preventDefault()}
          >
            Updated
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
        <OptionList.Separator />
        <OptionList.RadioGroup value={sortbyTime} onValueChange={setSortbyTime}>
          <OptionList.RadioGroupItem
            showIndicator={false}
            value="oldest"
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowUp size={16} />
            Oldest first
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="newest"
            showIndicator={false}
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowDown size={16} />
            Newest first
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
      </OptionList.Content>
    </OptionList.Root>
  );
};

export default Tools;
