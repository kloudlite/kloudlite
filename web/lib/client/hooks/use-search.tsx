import { useLocation, useNavigate } from '@remix-run/react';
import { MapType } from '../../types/common';

export interface IQueryParams {
  search?: string;
  page?: string;
}

export const encodeUrl = (values: MapType) => {
  return btoa(JSON.stringify(values || '{}'));
};

export const decodeUrl = (values: string | null = btoa('{}')) => {
  return JSON.parse(values ? atob(values) : '{}');
};

export const useQueryParameters = () => {
  const location = useLocation();
  const navigate = useNavigate();

  function setQueryParameters(params: IQueryParams) {
    const searchParams = new URLSearchParams(location.search);

    // Loop through the params object and set each key/value pair
    const entries = Object.entries(params);
    for (let i = 0; i < entries.length; i += 1) {
      const key = entries[i][0];
      const value = entries[i][1];
      searchParams.set(key, value);
    }

    // navigate({ ...location, search: searchParams.toString() });
    navigate(`${location.pathname}?${searchParams.toString()}`, {
      replace: true,
      state: {},
    });
  }

  function deleteQueryParameters(keys: string[]) {
    const searchParams = new URLSearchParams(location.search);

    // Loop through the params object and set each key/value pair
    for (let i = 0; i < keys.length; i += 1) {
      const key = keys[i];
      searchParams.delete(key);
    }

    // navigate({ ...location, search: searchParams.toString() });
    navigate(`${location.pathname}?${searchParams.toString()}`, {
      replace: true,
      state: {},
    });
  }

  return { setQueryParameters, deleteQueryParameters };
};
