import { Spinner } from '@jengaicons/react';
import { Await, useNavigate } from '@remix-run/react';
import { motion } from 'framer-motion';
import { ReactNode, Suspense, useEffect, useState } from 'react';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { parseError } from '~/root/lib/types/common';
import { MapType } from '../server/types/common';

interface SetTrueProps {
  setLoaded: (isLoaded: boolean) => void;
}

const SetTrue = ({ setLoaded }: SetTrueProps) => {
  useEffect(() => {
    setLoaded(true);
  }, []);
  return null;
};

interface SetCookieProps {
  _cookie: MapType[];
}

const SetCookie = ({ _cookie }: SetCookieProps) => {
  useEffect(() => {
    if (_cookie) {
      const cookie = getCookie();
      _cookie.forEach(({ name, value }) => {
        cookie.set(name, value);
      });
    }
  }, []);
  return null;
};

interface RedirectToProps {
  redirect: string;
}

const RedirectTo = ({ redirect }: RedirectToProps) => {
  const navigate = useNavigate();
  useEffect(() => {
    if (redirect) {
      navigate(redirect);
    }
  }, [redirect]);
  return null;
};

const GetSkeleton = ({
  skeleton = null,
  setLoaded = (_: boolean) => _,
}: any) => {
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

type LoadingDataType = any;

interface AwaitRespProps {
  data: LoadingDataType;
  error: string;
  redirect: string;
  cookie: MapType[];
  sample: Array<string>;
}

interface LoadingCompProps {
  data: Awaited<AwaitRespProps>;
  children?: (value: LoadingDataType) => ReactNode;
  skeleton?: ReactNode;
  errorComp?: ReactNode;
}

// NodesProps<string>

interface NodesProps<T> {
  nodes: T[];
  extra: string;
}

// const abc: <T>(arg: T) => T = (arg) => arg;
// const k: number = abc<number>(2);

export const LoadingComp = ({
  data,
  children = (_) => null,
  skeleton = null,
  errorComp = null,
}: LoadingCompProps) => {
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
          {(d) => {
            console.log(d.redirect);
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
                  >
                    <div className="flex flex-col max-h-[80vh] w-full bg-surface-basic-input border border-surface-basic-pressed on my-4xl rounded-md p-4xl gap-xl overflow-hidden">
                      <div className="font-bold text-xl text-[#A71B1B]">
                        Server Side Error:
                      </div>
                      <div className="flex overflow-scroll">
                        <div className="bg-[#A71B1B] w-2xl" />
                        <div className="overflow-auto max-h-full p-2xl flex-1 flex bg-[#EBEBEB] text-[#640C0C]">
                          <code>
                            {typeof d.error === 'string'
                              ? d.error
                              : JSON.stringify(d.error, null, 2)}
                          </code>
                        </div>
                      </div>
                    </div>
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

type pwTypes = <T>(fn: () => Promise<T>) => Promise<T | { error: string }>;

export const pWrapper: pwTypes = async (fn) => {
  try {
    return await fn();
  } catch (err) {
    return { error: parseError(err).message };
  }
};
