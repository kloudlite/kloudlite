import { useEffect, useState } from 'react';
import { FaDocker, FaGithub, FaGitlab } from 'react-icons/fa';
import { FiChevronDown, FiCopy, FiSearch } from 'react-icons/fi';
import { useOutletContext } from 'remix';
import { toast } from 'react-toastify';
import Skeleton from 'react-loading-skeleton';
import CopyToClipboard from 'react-copy-to-clipboard';
import classNames from 'classnames';
import { AnimatePresence, motion } from 'framer-motion';
import Button from '../components/atoms/buttons';
import { Input } from '../components/atoms/input';
import { Logo } from '../components/atoms/svg/logo';
import { Dialog } from '../components/elements/dialog';
import FormLayout from '../components/layouts/form-layout';
import { useAPIClient } from '../hooks/api-provider';
import logger from './log';
import { ListBlock } from '../components/elements/list-block';
import BounceIt from '../components/atoms/bounce-it';

export const isKlRegistryImage = ({ image }) => {
  if (!image) return false;
  const reg = /^[a-zA-Z.0-9]*/;

  const res = image.match(reg);
  if (!res || (res && res.length === 0)) return false;

  switch (res[0]) {
    case 'registry.kloudlite.io':
      return true;
    default:
      return false;
  }
};

export const getProviderIcon = ({ inputValue, isKl = false }) => {
  if (!inputValue) return null;
  if (isKl)
    return (
      <div className="w-4 h-4">
        <Logo />
      </div>
    );
  const reg = /^[a-zA-Z.0-9]*/;

  const res = inputValue.match(reg);
  if (!res || (res && res.length === 0)) return null;

  const getIcon = () => {
    switch (res[0]) {
      case 'registry.kloudlite.io':
        return (
          <div className="w-4 h-4">
            <Logo />
          </div>
        );
      case 'registry.gitlab.com':
        return (
          <div>
            <FaGitlab />
          </div>
        );
      case 'registry.github.com':
        return (
          <div>
            <FaGithub />
          </div>
        );
      default:
        return (
          <div>
            <FaDocker />
          </div>
        );
    }
  };
  return <span className="text-secondary-600 text-lg">{getIcon()}</span>;
};

export const trimImage = ({ image }) => {
  const reg = /\/.*$/;

  const res = image.match(reg);
  if (res && res.length > 0) {
    return res[0].replace(`/`, '');
  }

  return image;
};

export const trimImageName = ({ image }) => {
  const reg = /\/[a-zA-Z0-9_:]*$/;

  const res = image.match(reg);
  if (res && res.length > 0) {
    let val = res[0];
    val = val.replace('/', '');
    val = val.split(':');
    return val[0];
  }

  return image;
};

const ImageItem = ({
  imageName,
  handleChange,
  setOpen,
  setOpenImage,
  openImage,
}) => {
  const [tags, seTags] = useState([]);
  const api = useAPIClient();
  const [isLoading, setIsLoading] = useState(false);
  useEffect(() => {
    if (isLoading || openImage !== imageName || tags.length > 0) return;
    (async () => {
      setIsLoading(true);
      try {
        const { data, errors } = await api.getImageTags({
          imageName,
        });
        if (errors) {
          throw errors[0];
        }
        seTags(data);
      } catch (err) {
        toast.error(err.message);
        console.error(err.message);
      } finally {
        setIsLoading(false);
      }
    })();
  }, [openImage]);
  const randomIntFromInterval = (min, max) =>
    Math.floor(Math.random() * (max - min + 1) + min);
  function rangeBetwee(start, end) {
    const resarrult = [];
    if (start > end) {
      const arr = new Array(start - end + 1);
      // eslint-disable-next-line no-plusplus, no-param-reassign
      for (let i = 0; i < arr.length; i++, start--) {
        resarrult[i] = start;
      }
      return arr;
    }

    const arro = new Array(end - start + 1);

    // eslint-disable-next-line no-plusplus, no-param-reassign
    for (let j = 0; j < arro.length; j++, start++) {
      arro[j] = start;
    }
    return arro;
  }
  return (
    <ListBlock>
      <ListBlock.Content className={['px-6 py-3 flex-1', 'px-6 py-3']}>
        <ListBlock.Item
          noPadding
          cursorPointer
          key={imageName}
          onClick={() =>
            setOpenImage((s) => {
              if (s === imageName) return '';
              return imageName;
            })
          }
        >
          <div className="flex flex-col gap-0.5 justify-center">
            <span className="tracking-wider font-semibold text-neutral-600 select-none">
              {trimImage({ image: imageName })}
            </span>
            <span
              title={`${imageName}`}
              className="truncate text-xs max-w-[30rem] flex items-center gap-2"
            >
              <span className="text-neutral-600 tracking-wide font-medium truncate">
                {trimImageName({ image: imageName })}
              </span>
              <span onClick={(e) => e.stopPropagation()}>
                <CopyToClipboard
                  text={`registry.kloudlite.io/${imageName}`}
                  onCopy={() => {
                    toast.info('Image link copied to clipboard...');
                  }}
                >
                  <BounceIt className="cursor-pointer p-[0.25rem] rounded-full hover:bg-primary-100 hover:text-primary-900 transition-all">
                    <FiCopy />
                  </BounceIt>
                </CopyToClipboard>
              </span>
            </span>
          </div>

          <div className="flex gap-4 items-center">
            <BounceIt className="rounded-full border border-neutral-200 p-0.5 hover:bg-neutral-50 cursor-pointer">
              <FiChevronDown
                className={classNames('transition-all', {
                  'rotate-180': openImage === imageName,
                })}
              />
            </BounceIt>
          </div>
        </ListBlock.Item>
        <AnimatePresence>
          {openImage === imageName && (
            <motion.div
              initial={{ height: 0 }}
              animate={{ height: 'auto' }}
              exit={{ height: 0 }}
              className="overflow-hidden"
              transition={{
                ease: 'anticipate',
              }}
            >
              <ListBlock.Item>
                <div className="flex flex-col gap-2">
                  <span className="font-medium text-xs">Tags</span>
                  <div className="flex flex-wrap gap-2 items-center">
                    {isLoading &&
                      rangeBetwee(1, randomIntFromInterval(2, 3)).map((i) => {
                        return (
                          <Skeleton
                            key={i}
                            width={100}
                            height={25}
                            borderRadius={100}
                          />
                        );
                      })}
                    {!isLoading &&
                      tags.map(({ name }) => {
                        return (
                          <BounceIt
                            onClick={() => {
                              handleChange('image')({
                                target: {
                                  value: `registry.kloudlite.io/${imageName}:${name}`,
                                },
                              });
                              setOpen(false);
                            }}
                            className="border border-orange-600 rounded-full px-3 py-0.5 flex gap-1 items-center text-xs cursor-pointer"
                            key={name}
                          >
                            <span className="text-orange-600">{name}</span>
                          </BounceIt>
                        );
                      })}
                  </div>
                </div>
              </ListBlock.Item>
            </motion.div>
          )}
        </AnimatePresence>
      </ListBlock.Content>
    </ListBlock>
  );
};

export const RegistryImageInput = ({ values, handleChange }) => {
  // @ts-ignore
  const { rootContext } = useOutletContext();
  const { account } = rootContext || {};
  const { id: accountId } = account || {};

  const [isOpen, setOpen] = useState(false);

  const [inSearchString, setInSearchString] = useState('');
  const [searchString, setSearchString] = useState('');
  const [searchStringTimer, setSearchStringTimer] = useState(0);

  useEffect(() => {
    if (searchStringTimer) {
      clearTimeout(searchStringTimer);
    }
    setSearchStringTimer(
      // @ts-ignore
      setTimeout(() => {
        setSearchStringTimer(null);
        setSearchString(inSearchString);
      }, 500)
    );
  }, [inSearchString]);
  const api = useAPIClient();
  const [images, setImages] = useState([]);
  const [isLoading, setIsloading] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    (async () => {
      setIsloading(true);
      try {
        const { data, errors: searchErrors } = await api.harborSearch({
          accountId,
          q: searchString || 'acc',
        });
        if (searchErrors) {
          throw searchErrors[0];
        }
        setImages(data);
      } catch (err) {
        logger.error(err.message);
        toast.error(err.message);
      } finally {
        setIsloading(false);
      }
    })();
  }, [searchString, isOpen]);
  const [openImage, setOpenImage] = useState('');

  return (
    <>
      <Dialog {...{ isOpen, setOpen }}>
        <FormLayout
          heading="Select image"
          className="min-h-[50vh] max-h-[80vh]"
        >
          <Input
            value={inSearchString}
            onChange={(e) => setInSearchString(e.target.value)}
            className="min-w-[15rem]"
            icon={<FiSearch />}
            placeholder="Search for repo"
          />
          {isLoading && (
            <div className="gap-2 flex flex-col">
              {[1, 2, 3, 4, 5].map((i) => {
                return <Skeleton key={i} height={20} />;
              })}
            </div>
          )}
          {!isLoading && (
            <div className="flex flex-col gap-2">
              {/* <ListBlock className="overflow-y-auto overflow-x-hidden"> */}
              {/*   <ListBlock.Content */}
              {/*     className={['px-3 py-3 pl-4 flex-1', 'py-3 px-2 pr-4']} */}
              {/*   > */}
              {/*   </ListBlock.Content> */}
              {/* </ListBlock> */}

              {images.map(({ imageName }, i) => {
                if (i > 4) return null;

                return (
                  <ImageItem
                    key={imageName}
                    {...{
                      setOpen,
                      imageName,
                      handleChange,
                      setOpenImage,
                      openImage,
                    }}
                  />
                );
              })}
            </div>
          )}
        </FormLayout>
      </Dialog>
      <div className="flex items-end justify-between gap-3">
        <div className="flex flex-1 flex-col">
          <Input
            label="Image Url*"
            icon={getProviderIcon({ inputValue: values.image })}
            onChange={handleChange('image')}
            value={values.image}
            placeholder="Type Image Url here"
          />
        </div>

        <span className="font-semibold py-2">Or</span>

        <Button variant="primary-outline" onClick={() => setOpen(true)}>
          choose from registry
        </Button>
      </div>
    </>
  );
};
