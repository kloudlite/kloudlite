import { Spinner } from '@jengaicons/react';
import { SerializeFrom } from '@remix-run/node';
import { Await, useNavigate } from '@remix-run/react';
import { motion } from 'framer-motion';
import { ReactNode, Suspense, useEffect, useState } from 'react';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { FlatMapType, NN } from '~/root/lib/types/common';
import { parseError } from '~/root/lib/utils/common';

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
  _cookie: FlatMapType<string>[] | undefined;
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
      transition={{ ease: 'anticipate', duration: 0.1 }}
    >
      {skeleton || (
        <div className="pt-14xl flex items-center justify-center gap-2xl h-full">
          <span className="animate-spin">
            <Spinner color="currentColor" weight={2} size={28} />
          </span>
          <span className="text-[2rem]">Loading...</span>
        </div>
      )}
    </motion.div>
  );
};

interface AwaitRespProps {
  readonly error?: string;
  readonly redirect?: string;
  readonly cookie?: FlatMapType<string>[];
}

export type BaseData<T = any> = Promise<Awaited<AwaitRespProps & T>>;

interface LoadingCompProps<T = any> {
  data:
    | Promise<SerializeFrom<AwaitRespProps & T>>
    | SerializeFrom<AwaitRespProps & T>;
  children?: (value: NN<T>) => ReactNode;
  skeleton?: ReactNode;
  errorComp?: ReactNode;
}

export function LoadingComp<T>({
  data,
  children = (_) => null,
  skeleton = null,
  errorComp = null,
}: LoadingCompProps<T>) {
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
                    <div className="flex flex-col bg-surface-basic-input border border-surface-basic-pressed on my-4xl rounded-md p-4xl gap-xl">
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
                  className="relative loading-container"
                >
                  {children(d as any)}
                </motion.div>
              </>
            );
          }}
        </Await>
      </Suspense>
    </>
  );
}

type pwTypes = <T>(fn: () => Promise<T>) => Promise<T & AwaitRespProps>;

// @ts-ignore
export const pWrapper: pwTypes = async (fn) => {
  try {
    return await fn();
  } catch (err) {
    return { error: parseError(err).message };
  }
};
