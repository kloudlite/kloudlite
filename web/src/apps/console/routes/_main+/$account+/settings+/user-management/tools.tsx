import React, { useEffect, useState } from 'react';
import OptionList from '@kloudlite/design-system/atoms/option-list';
import Toolbar from '@kloudlite/design-system/atoms/toolbar';
import { ArrowDown, ArrowsDownUp, ArrowUp } from '~/console/components/icons';

interface Props {
  sortFunction: (options: {
    sortByProperty: string;
    sortByTime: string;
  }) => void;
}

const SortbyOptionList: React.FC<Props> = ({ sortFunction }) => {
  const [sortbyProperty, setSortbyProperty] = useState('updated');
  const [sortbyTime, setSortbyTime] = useState('des');

  useEffect(() => {
    sortFunction({ sortByProperty: sortbyProperty, sortByTime: sortbyTime });
  }, [sortbyProperty, sortbyTime]);

  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            <Toolbar.IconButton icon={<ArrowsDownUp />} variant="basic" />
          </div>
          <div className="flex md:hidden">
            <Toolbar.IconButton variant="basic" icon={<ArrowUp />} />
          </div>
        </div>
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.RadioGroup
          value={sortbyProperty}
          onValueChange={setSortbyProperty}
        >
          <OptionList.RadioGroupItem value="name">
            Name
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem value="updated">
            Updated
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
        <OptionList.Separator />
        <OptionList.RadioGroup value={sortbyTime} onValueChange={setSortbyTime}>
          <OptionList.RadioGroupItem showIndicator={false} value="asc">
            <ArrowUp size={16} />
            Ascending
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem showIndicator={false} value="des">
            <ArrowDown size={16} />
            Descending
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
      </OptionList.Content>
    </OptionList.Root>
  );
};

const Tools = ({
  searchText,
  setSearchText,
  sortTeamMembers,
}: {
  searchText: string;
  setSearchText: (s: string) => void;
  sortTeamMembers: (options: {
    sortByProperty: string;
    sortByTime: string;
  }) => void;
}) => {
  return (
    <div className="flex-1 ">
      <Toolbar.Root>
        <div className="w-full flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <Toolbar.TextInput
              value={searchText}
              onChange={(v) => {
                setSearchText(v.target.value);
              }}
              placeholder="Search"
            />
          </div>
          <SortbyOptionList sortFunction={sortTeamMembers} />
        </div>
      </Toolbar.Root>
    </div>
  );
};
export default Tools;
