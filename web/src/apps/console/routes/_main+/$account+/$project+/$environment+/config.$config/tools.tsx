import React from 'react';
import Toolbar from '~/components/atoms/toolbar';
import ViewMode from '~/console/components/view-mode';

const Tools = ({
  searchText,
  setSearchText,
}: {
  searchText: string;
  setSearchText: React.Dispatch<React.SetStateAction<string>>;
}) => {
  return (
    <div className="mb-6xl">
      <Toolbar.Root>
        <div className="flex-1">
          <Toolbar.TextInput
            placeholder="Search"
            value={searchText}
            onChange={({ target }) => {
              setSearchText(target.value);
            }}
          />
        </div>
        <ViewMode />
      </Toolbar.Root>
    </div>
  );
};

export default Tools;
