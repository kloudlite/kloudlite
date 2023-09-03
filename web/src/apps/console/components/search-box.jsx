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
import { isValidRegex } from '../server/r-urils/common';

export const SearchBox = ({ InputElement = Toolbar.TextInput }) => {
  const [searchParams] = useSearchParams();

  const searchObject = decodeUrl(searchParams.get('search'));

  const [search, setSearch] = useState(() => searchObject?.text?.regex || '');

  const { setQueryParameters } = useQueryParameters();
  const [isFirstTime, setIsFirstTime] = useState(true);

  useDebounce(
    () => {
      if (isFirstTime) {
        setIsFirstTime(false);
        return;
      }
      if (search) {
        if (isValidRegex(search)) {
          setQueryParameters({
            search: encodeUrl({
              ...searchObject,
              text: {
                matchType: 'regex',
                regex: search,
              },
            }),
          });
        }
      } else if (searchObject?.text) {
        delete searchObject.text;
        setQueryParameters({
          search: encodeUrl(searchObject),
        });
      }
    },
    300,
    [search]
  );

  return (
    <div className="w-full">
      <InputElement
        value={search}
        onChange={(e) => {
          setSearch(e.target.value);
        }}
        placeholder="Search"
        prefixIcon={<Search/>}
      />
    </div>
  );
};
