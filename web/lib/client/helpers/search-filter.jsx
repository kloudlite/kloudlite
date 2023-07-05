import Fuse from 'fuse.js';
import { useCallback, useState } from 'react';

export const searchFilter = ({
  data,
  searchText,
  reverse = false,
  keys = [],
  remainOrder = false,
}) => {
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
    // findAllMatches: true,
    threshold: 0.0,
    // distance: 300,
    useExtendedSearch: true,
    includeMatches: true,
    ignoreLocation: true,
    shouldSort: !remainOrder,
  });

  const results = fuse.search(searchText);

  return results.map(({ item, ...etc }) => ({
    ...item,
    searchInf: etc,
  }));
};

export const useSearch = (
  { data, searchText, reverse = false, keys = [], remainOrder = false },
  dependency = []
) => {
  return useCallback(
    () =>
      searchFilter({
        data,
        searchText,
        reverse,
        keys,
        remainOrder,
      }),
    dependency
  )();
};

export const useInputSearch = (
  { data, reverse = false, keys = [] },
  dependency = []
) => {
  const [searchText, setSearchText] = useState('');
  return [
    {
      value: searchText,
      onChange: (e) => setSearchText(e.target.value),
    },
    useCallback(
      () =>
        searchFilter({
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
