import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useSearchParams } from '@remix-run/react';
import Toolbar from '~/components/atoms/toolbar';
import { useState } from 'react';
import { Search } from '@jengaicons/react';

export const SearchBox = ({ fields = ['metadata.name'] }) => {
  const [sp] = useSearchParams();

  const [search, setSearch] = useState(
    () => decodeUrl(sp.get('search')).keyword || ''
  );
  const { setQueryParameters, deleteQueryParameters } = useQueryParameters();
  const [isFirstTime, setIsFirstTime] = useState(true);

  useDebounce(search, 300, () => {
    if (isFirstTime) {
      setIsFirstTime(false);
      return;
    }
    if (search) {
      setQueryParameters({
        search: encodeUrl({
          fields,
          keyword: search || '',
        }),
      });
    } else if (decodeUrl(sp.get('search')).keyword || '') {
      deleteQueryParameters(['search']);
    }
  });

  return (
    <div className="w-full">
      <Toolbar.TextInput
        value={search}
        onChange={(e) => {
          setSearch(e.target.value);
        }}
        placeholder="Search"
        prefixIcon={Search}
      />
    </div>
  );
};
