import { Await, useNavigate } from '@remix-run/react';
import { Suspense, useEffect, useState } from 'react';
import { getCookie } from '~/root/lib/app-setup/cookies';

export const LoadingComp = ({
  data,
  children = (_) => null,
  skeleton = null,
  errorComp = null,
}) => {
  const [redirect, setRed] = useState('');
  const [_cookie, setCookie] = useState('');
  const navigate = useNavigate();
  useEffect(() => {
    if (redirect) {
      navigate(redirect);
    }
  }, [redirect]);

  useEffect(() => {
    if (_cookie) {
      const cookie = getCookie();
      _cookie.forEach(({ name, value }) => {
        cookie.set(name, value);
      });
    }
  }, [_cookie]);

  if (typeof children !== 'function') {
    return children;
  }

  return (
    <Suspense fallback={skeleton || <div>Loading...</div>}>
      <Await
        resolve={data}
        errorElement={errorComp || <div>Something Went Wrong</div>}
      >
        {(d = {}) => {
          if (d.cookie) {
            setCookie(d.cookie);
          }
          if (d.redirect) {
            setRed(d.redirect);
            return null;
          }
          if (d.error) {
            return (
              <pre className="text-text-critical">
                <code>{JSON.stringify(d.error, null, 2)}</code>
              </pre>
            );
          }
          return children(d);
        }}
      </Await>
    </Suspense>
  );
};
export const pWrapper = async (fn) => {
  try {
    return await fn();
  } catch (err) {
    return { error: err.message };
  }
};
