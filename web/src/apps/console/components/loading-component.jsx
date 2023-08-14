import { Spinner } from '@jengaicons/react';
import { Await, useNavigate } from '@remix-run/react';
import { motion } from 'framer-motion';
import { Suspense, useEffect, useState } from 'react';
import { getCookie } from '~/root/lib/app-setup/cookies';

const SetTrue = ({ setLoaded = (_) => _ }) => {
  useEffect(() => {
    setLoaded(true);
  }, []);
  return null;
};

const SetCookie = ({ _cookie }) => {
  useEffect(() => {
    if (_cookie) {
      const cookie = getCookie();
      // @ts-ignore
      _cookie.forEach(({ name, value }) => {
        cookie.set(name, value);
      });
    }
  }, []);
  return null;
};

const RedirectTo = ({ redirect }) => {
  const navigate = useNavigate();
  useEffect(() => {
    if (redirect) {
      navigate(redirect);
    }
  }, [redirect]);
  return null;
};

const GetSkeleton = ({ skeleton = null, setLoaded = (_) => _ }) => {
  useEffect(() => {
    setLoaded(true);
  }, []);
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ ease: 'anticipate' }}
    >
      {skeleton || (
        <div className="pt-14xl flex items-center justify-center gap-2xl h-full">
          <motion.span
            initial={{ width: 0 }}
            animate={{ width: 'auto', paddingRight: 0 }}
            exit={{ width: 0 }}
            className="flex items-center justify-center aspect-square overflow-hidden"
          >
            <span className="animate-spin">
              <Spinner color="currentColor" weight={2} size={24} />
            </span>
          </motion.span>
          <span className="text-[2rem]">Loading...</span>
        </div>
      )}
    </motion.div>
  );
};

export const LoadingComp = ({
  data,
  children = (_) => null,
  skeleton = null,
  errorComp = null,
}) => {
  const [skLoaded, setSkLoaded] = useState(false);

  if (typeof children !== 'function') {
    return children;
  }

  return (
    <>
      {!skLoaded && <GetSkeleton skeleton={skeleton} />}

      <Suspense
        fallback={<GetSkeleton skeleton={skeleton} setLoaded={setSkLoaded} />}
      >
        <Await
          resolve={data}
          errorElement={errorComp || <div>Something Went Wrong</div>}
        >
          {(d = {}) => {
            if (d.redirect) {
              return (
                <>
                  <SetTrue setLoaded={setSkLoaded} />
                  <SetCookie _cookie={d.cookie} />
                  <RedirectTo redirect={d.redirect} />
                </>
              );
            }
            if (d.error) {
              return (
                <>
                  <SetTrue setLoaded={setSkLoaded} />
                  <SetCookie _cookie={d.cookie} />
                  <motion.pre
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ ease: 'anticipate' }}
                    className="text-text-critical"
                  >
                    <code>
                      Error:{' '}
                      {typeof d.error === 'string'
                        ? d.error
                        : JSON.stringify(d.error, null, 2)}
                    </code>
                  </motion.pre>
                </>
              );
            }
            return (
              <>
                <SetTrue setLoaded={setSkLoaded} />
                <SetCookie _cookie={d.cookie} />
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ ease: 'anticipate' }}
                >
                  {children(d)}
                </motion.div>
              </>
            );
          }}
        </Await>
      </Suspense>
    </>
  );
};
export const pWrapper = async (fn = async (_) => _) => {
  try {
    return await fn();
  } catch (err) {
    return { error: err.message };
  }
};
