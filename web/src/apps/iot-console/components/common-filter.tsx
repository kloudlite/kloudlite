import { CaretDownFill, Search } from '@jengaicons/react';
import { useSearchParams } from '@remix-run/react';
import { useState, Key } from 'react';
import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';
import { handleError } from '~/root/lib/utils/common';
import { FilterType, IdataFetcher } from './filters';

interface IOnCheckHandler {
  searchParams: URLSearchParams;
  type: string;
  check: boolean;
  setQueryParameters: (v: any) => void;
  value: string | boolean;
}

export const onCheckHandler = ({
  searchParams,
  type,
  check,
  setQueryParameters,
  value,
}: IOnCheckHandler) => {
  const search = decodeUrl(searchParams.get('search'));
  const array: (string | boolean)[] = search?.[type]?.array || [];

  let nArray: (string | boolean)[] = [];
  if (!check) {
    nArray = array.filter((_v) => {
      return _v !== value;
    });
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

interface IdataFormer {
  data?: { content: string; value: string | boolean }[];
  searchParams: URLSearchParams;
  type: string;
}

const dataFormer = ({
  data = [{ content: '', value: '' }],
  searchParams,
  type,
}: IdataFormer) => {
  const array = decodeUrl(searchParams.get('search'))?.[type]?.array || [];

  return data.map(({ content, value }) => {
    return {
      checked: array.indexOf(value) !== -1,
      content,
      value,
    };
  });
};

interface IOptioniList {
  open?: boolean;
  setOpen?: React.Dispatch<React.SetStateAction<boolean>>;
  name: string;
  search: boolean;
  dataFetcher: IdataFetcher;
  type: string;
}

const OptioniList = ({
  open = false,
  setOpen = (_) => _,
  name,
  search,
  dataFetcher,
  type,
}: IOptioniList) => {
  const [options, setOptions] = useState<
    { content: string; checked: boolean; value: string | boolean }[]
  >([]);
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
          value={name}
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
            key={checkItem.value as Key}
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

interface ICommonFilterOptions {
  options: FilterType[];
}

export const CommonFilterOptions = ({ options }: ICommonFilterOptions) => {
  return (
    <Toolbar.ButtonGroup.Root>
      {options.map((o) => {
        const { name } = o;
        return <OptioniList key={name} {...o} />;
      })}
    </Toolbar.ButtonGroup.Root>
  );
};
