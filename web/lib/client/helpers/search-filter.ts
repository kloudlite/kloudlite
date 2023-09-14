import Fuse from 'fuse.js';
import { ChangeEvent, useCallback, useState } from 'react';

interface IsearchFilter<T> {
  data: T[];
  searchText: string;
  reverse?: boolean;
  keys?: string[];
  remainOrder?: boolean;
  threshold: number;
}

export interface ISearchInfProps {
  searchInf: {
    refIndex: number;
    matches?: readonly Fuse.FuseResultMatch[];
    score?: number;
  };
}

type ISearchResp<T> = (T & ISearchInfProps)[];

export const searchFilter = <T>({
  data,
  searchText,
  reverse = false,
  keys = [],
  remainOrder = false,
  threshold,
}: IsearchFilter<T>): ISearchResp<T> => {
  if (!searchText) {
    if (reverse) {
      return [...(data || [])].reverse().map((item, index) => ({
        ...item,
        searchInf: {
          refIndex: index,
          matches: [],
        },
      }));
    }
    return data.map((item, index) => ({
      ...item,
      searchInf: {
        refIndex: index,
        matches: [],
      },
    }));
  }

  const fuse = new Fuse(data || [], {
    keys,
    findAllMatches: true,
    threshold,
    // distance: 300,
    useExtendedSearch: true,
    includeMatches: true,
    ignoreLocation: true,
    shouldSort: !remainOrder,
    // sortFn: (a, b) => {
    //   console.log(a, b);
    //   return -1;
    // },
  });

  const results = fuse.search(searchText);

  return results.map(({ item, ...etc }) => ({
    ...item,
    searchInf: etc,
  }));
};

interface IuseSearch<T> {
  data: T[];
  searchText: string;
  reverse?: boolean;
  keys?: any[];
  remainOrder?: boolean;
  threshold?: number;
}

export const useSearch = <T>(
  {
    data,
    searchText,
    reverse = false,
    keys = [],
    remainOrder = false,
    threshold = 0.2,
  }: IuseSearch<T>,
  dependency: any[] = []
): ISearchResp<T> => {
  return useCallback(
    () =>
      searchFilter({
        threshold,
        data,
        searchText,
        reverse,
        keys,
        remainOrder,
      }),
    dependency
  )();
};

interface IuseInputSearch {
  data: any[];
  reverse: boolean;
  keys: any[];
  threshold?: number;
}

export const useInputSearch = (
  { data, reverse = false, keys = [], threshold = 0.2 }: IuseInputSearch,
  dependency = []
) => {
  const [searchText, setSearchText] = useState('');
  return [
    {
      value: searchText,
      onChange: (e: ChangeEvent<HTMLInputElement>) =>
        setSearchText(e.target.value),
    },
    useCallback(
      () =>
        searchFilter({
          threshold,
          data,
          searchText,
          reverse,
          keys,
        }),
      [...dependency, searchText]
    )(),
    searchText,
  ];
};
