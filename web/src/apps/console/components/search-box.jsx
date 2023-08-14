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

const isValidRegex = (regexString = '') => {
  let isValid = true;
  try {
    // eslint-disable-next-line no-new
    new RegExp(regexString);
  } catch (e) {
    isValid = false;
  }
  return isValid;
};

export const SearchBox = ({
  // @ts-ignore
  InputElement = Toolbar.TextInput,
}) => {
  const [sp] = useSearchParams();

  const [search, setSearch] = useState(
    () => decodeUrl(sp.get('search'))?.text?.exact || ''
  );
  const { setQueryParameters, deleteQueryParameters } = useQueryParameters();
  const [isFirstTime, setIsFirstTime] = useState(true);

  useDebounce(search, 300, () => {
    if (isFirstTime) {
      setIsFirstTime(false);
      return;
    }
    if (search) {
      if (isValidRegex(search)) {
        setQueryParameters({
          search: encodeUrl({
            text: {
              matchType: 'regex',
              regex: search,
            },
          }),
        });
      }
    } else if (decodeUrl(sp.get('search'))?.text?.regex || '') {
      deleteQueryParameters(['search']);
    }
  });

  return (
    <div className="w-full">
      <InputElement
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
