import { useLocation, useNavigate } from '@remix-run/react';

export const encodeUrl = (values) => {
  return btoa(JSON.stringify(values || '{}'));
};

export const decodeUrl = (values) => {
  return JSON.parse(values ? atob(values) : '{}');
};

export const useQueryParameters = () => {
  const location = useLocation();
  const navigate = useNavigate();

  function setQueryParameters(params) {
    const searchParams = new URLSearchParams(location.search);

    // Loop through the params object and set each key/value pair
    const keys = Object.keys(params);
    for (let i = 0; i < keys.length; i += 1) {
      const key = keys[i];
      searchParams.set(key, params[key]);
    }

    // navigate({ ...location, search: searchParams.toString() });
    navigate(`${location.pathname}?${searchParams.toString()}`, {
      replace: true,
      state: {},
    });
  }

  function deleteQueryParameters(keys) {
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
