import { CaretDownFill, Search } from '@jengaicons/react';
import { useSearchParams } from '@remix-run/react';
import { useState } from 'react';
import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';
import { handleError } from '~/root/lib/utils/common';

export const onCheckHandler = ({
  searchParams,
  type,
  check,
  setQueryParameters,
  value,
}) => {
  const search = decodeUrl(searchParams.get('search'));
  const array = search?.[type]?.array || [];

  let nArray = [];
  if (!check) {
    nArray = array.filter((_v) => _v !== value);
  } else {
    nArray = [...array, value];
  }

  if (nArray.length === 0 && !!search[type]) {
    delete search[type];
    setQueryParameters({
      search: encodeUrl(search),
    });
  } else {
    setQueryParameters({
      search: encodeUrl({
        ...search,
        [type]: {
          matchType: 'array',
          array: nArray,
        },
      }),
    });
  }
};

const dataFormer = ({
  data = [{ content: '', value: '' }],
  searchParams,
  type,
}) => {
  const array = decodeUrl(searchParams.get('search'))?.[type]?.array || [];

  return data.map(({ content, value }) => {
    return {
      checked: array.indexOf(value) !== -1,
      content,
      value,
    };
  });
};

export const CommonFilterOptions = ({ options }) => {
  return (
    <Toolbar.ButtonGroup.Root>
      {options.map((o) => {
        const { name } = o;
        return <OptioniList key={name} {...o} />;
      })}
    </Toolbar.ButtonGroup.Root>
  );
};

const OptioniList = ({
  open = false,
  setOpen = (_) => _,
  name,
  search,
  dataFetcher,
  type,
}) => {
  const [options, setOptions] = useState([]);
  const [searchText, setSearchText] = useState('');

  const [isLoading, setLoading] = useState(false);
  const [searchParams] = useSearchParams();
  const { setQueryParameters } = useQueryParameters();

  useDebounce(
    async () => {
      if (isLoading) {
        return;
      }
      try {
        setLoading(true);
        const _data = await dataFetcher(searchText);
        const res = dataFormer({ data: _data, searchParams, type });
        setOptions(res);
      } catch (err) {
        handleError(err);
      } finally {
        setLoading(false);
      }
    },
    300,
    [searchText, searchParams]
  );

  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content={name}
          variant="basic"
          suffix={<CaretDownFill />}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        {search && (
          <OptionList.TextInput
            value={searchText}
            onChange={(e) => {
              setSearchText(e.target.value);
            }}
            placeholder="Filter cluster"
            prefixIcon={<Search />}
          />
        )}
        {isLoading && <OptionList.Item> Loading... </OptionList.Item>}
        {options.map((checkItem) => (
          <OptionList.CheckboxItem
            key={checkItem.value}
            checked={checkItem.checked}
            onValueChange={(value) => {
              onCheckHandler({
                value: checkItem.value,
                check: value,
                searchParams,
                setQueryParameters,
                type,
              });
            }}
            onClick={(e) => e.preventDefault()}
          >
            {checkItem.content}
          </OptionList.CheckboxItem>
        ))}
      </OptionList.Content>
    </OptionList.Root>
  );
};
