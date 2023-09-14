/* eslint-disable no-nested-ternary */
import Fuse from 'fuse.js';
import { ArrowsIn, ArrowsOut, List } from '@jengaicons/react';
import Anser from 'anser';
import axios from 'axios';
import classNames from 'classnames';
import hljs from 'highlight.js';
import React, {
  ReactNode,
  useCallback,
  useEffect,
  useRef,
  useState,
} from 'react';
import { ViewportList } from 'react-viewport-list';
import * as sock from 'websocket';
import {
  ISearchInfProps,
  useSearch,
} from '~/root/lib/client/helpers/search-filter';
import useClass from '~/root/lib/client/hooks/use-class';
import { dayjs } from '~/components/molecule/dayjs';

type ILog = { message: string; timestamp: string };
type ILogWithPodName = ILog & { pod_name: string; lineNumber: number };

type ISocketMessage = {
  pod_name: string;
  logs: ILog[];
};

const padLeadingZeros = (num: number, size: number) => {
  let s = `${num}`;
  while (s.length < size) s = `0${s}`;
  return s;
};

const threshold = 0.4;

interface IHighlightIt {
  language: string;
  inlineData: string;
  className?: string;
  enableHL?: boolean;
}

const HighlightIt = ({
  language,
  inlineData = '',
  className = '',
  enableHL = false,
}: IHighlightIt) => {
  const ref = useRef(null);
  const data = Anser.ansiToText(inlineData);

  useEffect(() => {
    (async () => {
      if (ref.current) {
        if (enableHL) {
          // if (!isScrolledIntoView(ref.current)) return;
          // @ts-ignore
          ref.current.innerHTML = hljs.highlight(
            data,
            {
              language,
            },
            false
          ).value;
        } else {
          // @ts-ignore
          ref.current.innerHTML = Anser.ansiToHtml(inlineData);
        }

        // @ts-ignore
      }
    })();
  }, [inlineData, language]);

  return (
    <div ref={ref} className={classNames(className, 'inline')}>
      {data}
    </div>
  );
};

interface ILineNumber {
  lineNumber: number;
  fontSize: number;
  lines: number;
}
const LineNumber = ({ lineNumber, fontSize, lines }: ILineNumber) => {
  const ref = useRef(null);
  const [data, setData] = useState(() => padLeadingZeros(1, `${lines}`.length));

  useEffect(() => {
    setData(padLeadingZeros(lineNumber, `${lines}`.length));
  }, [lines, lineNumber]);
  return (
    <code
      key={`ind+${lineNumber}`}
      className="inline-flex gap-xl items-center whitespace-pre"
      ref={ref}
    >
      <span className="flex sticky left-0" style={{ fontSize }}>
        <HighlightIt
          enableHL
          inlineData={data}
          language="accesslog"
          className="border-b border-border-tertiary bg-surface-tertiary-hovered px-md"
        />
        <div className="hljs" />
      </span>
    </code>
  );
};

interface IFilterdHighlightIt {
  searchInf?: ISearchInfProps['searchInf'];
  inlineData: string;
  className?: string;
  language: string;
  searchText: string;
  showAll: boolean;
}

interface HighlightProps {
  value: string;
  indices: Array<[number, number]>;
}

const Highlighter: React.FC<HighlightProps> = ({ value, indices }) => {
  let lastIndex = 0;
  const parts = [];

  indices.forEach(([start, end]) => {
    if (lastIndex !== start) {
      parts.push(
        <span style={{ opacity: 0.8 }} key={lastIndex}>
          <HighlightIt
            language="accesslog"
            inlineData={value.substring(lastIndex, start)}
          />
        </span>
      );
    }
    parts.push(<span key={start}>{value.substring(start, end + 1)}</span>);
    lastIndex = end + 1;
  });

  if (lastIndex !== value.length) {
    parts.push(<span key={lastIndex}>{value.substring(lastIndex)}</span>);
  }

  return parts;
};

const InlineSearch = ({
  inlineData = '',
  className = '',
  language,
  searchText,
}: IFilterdHighlightIt) => {
  const res = useSearch(
    {
      data: [{ message: inlineData }],
      keys: ['message'],
      searchText,
      threshold,
      remainOrder: true,
    },
    [inlineData, searchText]
  );

  if (res.length && res[0].searchInf.matches?.length) {
    const def: Fuse.RangeTuple[] = [];
    return (
      <Highlighter
        {...{
          value: inlineData,
          indices:
            res[0].searchInf.matches?.reduce((acc, curr) => {
              return [...acc, ...curr.indices];
            }, def) || def,
        }}
      />
    );
  }
  return (
    <HighlightIt
      {...{
        inlineData,
        language,
        className: classNames(className, {
          'opacity-20': !!searchText,
        }),
        enableHL: true,
      }}
    />
  );
};

const FilterdHighlightIt = ({
  searchInf,
  inlineData = '',
  className = '',
  language,
  searchText,
  showAll,
}: IFilterdHighlightIt) => {
  const def: Fuse.RangeTuple[] = [];

  if (showAll) {
    return (
      <div className={classNames('whitespace-pre', className)}>
        <InlineSearch
          {...{
            language,
            inlineData,
            searchText,
            className,
            showAll,
          }}
        />
      </div>
    );
  }

  return (
    <div className={classNames('whitespace-pre', className)}>
      {searchInf?.matches?.length ? (
        <Highlighter
          key={inlineData}
          {...{
            value: inlineData,
            indices: searchInf.matches.reduce((acc, curr) => {
              // const validIndices = curr.indices.filter((i) => {
              //   return i[1] - i[0] >= searchText.length - 1;
              // });
              // console.log(curr.indices, validIndices);
              return [...acc, ...curr.indices];
            }, def),
          }}
        />
      ) : (
        <HighlightIt
          {...{
            inlineData,
            language,
            enableHL: true,
          }}
        />
      )}
    </div>
  );
};

interface ILogLine {
  fontSize: number;
  selectableLines: boolean;
  showAll: boolean;
  searchText: string;
  language: string;
  lines: number;
  hideLines?: boolean;
  log: ILogWithPodName & {
    searchInf?: ISearchInfProps['searchInf'];
  };
}

const LogLine = ({
  log,
  fontSize,
  selectableLines,
  showAll,
  searchText,
  language,
  lines,
  hideLines,
}: ILogLine) => {
  return (
    <code
      className={classNames(
        'flex py-xs items-center whitespace-pre border-b border-transparent transition-all',
        {
          'hover:bg-surface-tertiary-hovered cursor-pointer': selectableLines,
        }
      )}
      style={{
        fontSize,
        paddingLeft: fontSize / 4,
        paddingRight: fontSize / 2,
      }}
    >
      {!hideLines && (
        <LineNumber
          lineNumber={log.lineNumber}
          lines={lines}
          fontSize={fontSize}
        />
      )}

      <div className="w-[3px] mr-xl bg-surface-success-default h-full" />
      <div className="mr-xl h-full">{log.pod_name}</div>
      <div className="inline-flex gap-xl">
        <HighlightIt
          {...{
            inlineData: `${dayjs(log.timestamp).format('lll')} |`,
            language: 'apache',
            enableHL: true,
          }}
        />

        <FilterdHighlightIt
          {...{
            searchText,
            inlineData: log.message,
            searchInf: log.searchInf,
            language,
            showAll,
          }}
        />
      </div>
    </code>
  );
};

interface ILogBlock {
  data: ISocketMessage[];
  maxLines?: number;
  follow: boolean;
  enableSearch: boolean;
  selectableLines: boolean;
  title: ReactNode;
  noScrollBar: boolean;
  fontSize: number;
  actionComponent: ReactNode;
  hideLines: boolean;
  language: string;
  solid: boolean;
}

const LogBlock = ({
  data = [],
  follow,
  enableSearch,
  selectableLines,
  title,
  noScrollBar,
  maxLines,
  fontSize,
  actionComponent,
  hideLines,
  language,
  solid,
}: ILogBlock) => {
  const [searchText, setSearchText] = useState('');

  const temp: { res: ILogWithPodName[]; id: number } = {
    res: [],
    id: 1,
  };

  const flatLogs = useCallback(
    () =>
      data.reduce((acc, curr) => {
        let { id } = acc;
        const tres = [
          ...acc.res,
          ...curr.logs.map((log, index) => {
            id = acc.id + index;
            return {
              ...log,
              pod_name: curr.pod_name,
              lineNumber: id,
            };
          }),
        ];

        return {
          id,
          res: tres,
        };
      }, temp).res,
    [data]
  )();

  const searchResult = useSearch(
    {
      data: flatLogs,
      keys: ['message'],
      searchText,
      threshold,
      remainOrder: true,
    },
    [data, searchText]
  );

  const [showAll, setShowAll] = useState(true);
  const ref = useRef(null);

  useEffect(() => {
    (async () => {
      if (follow && ref.current) {
        // @ts-ignore
        ref.current.scrollTo(0, ref.current.scrollHeight);
      }
    })();
  }, [data, maxLines]);

  return (
    <div
      className={classNames('hljs p-xs flex flex-col gap-sm h-full', {
        'rounded-md': !solid,
      })}
    >
      <div className="flex justify-between items-center border-b border-border-tertiary p-lg">
        <div className="">{data ? title : 'No logs found'}</div>

        <div className="flex items-center gap-xl">
          {actionComponent}
          {enableSearch && (
            <form
              className="flex gap-xl items-center text-sm"
              onSubmit={(e) => {
                e.preventDefault();
                setShowAll((s) => !s);
              }}
            >
              <input
                className="bg-transparent border border-surface-tertiary-default rounded-md px-lg py-xs w-[10rem]"
                placeholder="Search"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
              />
              <div
                onClick={() => {
                  setShowAll((s) => !s);
                }}
                className="cursor-pointer active:translate-y-[1px] transition-all"
              >
                <span
                  className={classNames('font-medium', {
                    'opacity-50': showAll,
                    'text-text-secondary': !showAll,
                  })}
                >
                  <List color="currentColor" size={16} />
                </span>
              </div>
              <code className={classNames('text-xs font-bold', {})}>
                {searchResult.length}
                matches
              </code>
            </form>
          )}
        </div>
      </div>

      <div
        className={classNames('flex flex-1 overflow-auto', {
          'no-scroll-bar': noScrollBar,
          'hljs-log-scrollbar': !noScrollBar,
        })}
      >
        <div className="flex flex-1 h-full">
          <div
            className="flex-1 flex flex-col pb-8"
            style={{ lineHeight: `${fontSize * 1.5}px` }}
            ref={ref}
          >
            <ViewportList items={showAll ? flatLogs : searchResult}>
              {(log) => {
                return (
                  <LogLine
                    log={log}
                    language={language}
                    searchText={searchText}
                    fontSize={fontSize}
                    lines={flatLogs.length}
                    showAll={showAll}
                    key={log.lineNumber}
                    hideLines={hideLines}
                    selectableLines={selectableLines}
                  />
                );
              }}
            </ViewportList>
          </div>
        </div>
      </div>
    </div>
  );
};

interface IHighlightJsLog {
  websocket?: boolean;
  follow?: boolean;
  url?: string;
  text?: string;
  enableSearch?: boolean;
  selectableLines?: boolean;
  title?: ReactNode;
  height?: string;
  width?: string;
  noScrollBar?: boolean;
  maxLines?: number;
  fontSize?: number;
  loadingComponent?: ReactNode;
  actionComponent?: ReactNode;
  hideLines?: boolean;
  language?: string;
  solid?: boolean;
  className?: string;
  dark?: boolean;
}

const HighlightJsLog = ({
  websocket = false,
  follow = true,
  url = '',
  enableSearch = true,
  selectableLines = true,
  title = '',
  height = '400px',
  width = '600px',
  noScrollBar = false,
  maxLines,
  fontSize = 14,
  loadingComponent = null,
  actionComponent = null,
  hideLines = false,
  language = 'accesslog',
  solid = false,
  className = '',
}: IHighlightJsLog) => {
  const [messages, setMessages] = useState<ISocketMessage[]>([]);
  const [errors, setErrors] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [fullScreen, setFullScreen] = useState(false);

  const { setClassName, removeClassName } = useClass({
    elementClass: 'loading-container',
  });

  useEffect(() => {
    (async () => {
      if (!url || websocket) return;
      setIsLoading(true);
      try {
        const d = await axios({
          url,
          method: 'GET',
        });
        setMessages((d.data || '').trim());
      } catch (err) {
        setErrors(
          `${(err as Error).message}
An error occurred attempting to load the provided log.
Please check the URL and ensure it is reachable.
${url}`
        );
      } finally {
        setIsLoading(false);
      }
    })();
  }, []);

  useEffect(() => {
    if (!url || !websocket) return () => {};

    let wsclient: sock.w3cwebsocket;
    setIsLoading(true);
    try {
      // eslint-disable-next-line new-cap
      wsclient = new sock.w3cwebsocket(url, '', '', {});
    } catch (err) {
      setIsLoading(false);
      setErrors(
        `${(err as Error).message}
An error occurred attempting to load the provided log.
Please check the URL and ensure it is reachable.
${url}`
      );
      return () => {};
    }
    // wsclient.onopen = logger.log;
    // wsclient.onclose = logger.log;
    // wsclient.onerror = logger.log;

    wsclient.onmessage = (msg: sock.IMessageEvent) => {
      try {
        const data: ISocketMessage[] = JSON.parse(msg.data.toString());

        setMessages((s) => [...s, ...data]);
        setIsLoading(false);
      } catch (err) {
        console.log(err);
        setErrors("'Something went wrong! Please try again.'");
      }
    };
    return () => {
      wsclient.close();
    };
  }, []);

  useEffect(() => {
    const keyDownListener = (e: any) => {
      if (e.code === 'Escape') {
        e.stopPropagation();
        setFullScreen(false);
      }
    };

    if (fullScreen && window?.document?.children[0]) {
      // @ts-ignore
      window.document.children[0].style = `overflow-y:hidden`;

      document.addEventListener('keydown', keyDownListener);
    } else if (window?.document?.children[0]) {
      // @ts-ignore
      window.document.children[0].style = `overflow-y:auto`;

      document.removeEventListener('keydown', keyDownListener);
    }
  }, [fullScreen]);

  return (
    <div
      className={classNames(className, {
        'fixed w-full h-full left-0 top-0 z-[999] bg-black': fullScreen,
      })}
      style={{
        width: fullScreen ? '100%' : width,
        height: fullScreen ? '100vh' : height,
      }}
    >
      {isLoading ? (
        loadingComponent || (
          <div className="hljs p-xs rounded-md flex flex-col gap-sm items-center justify-center h-full">
            <code className="">
              <HighlightIt language={language} inlineData="Loading..." />
            </code>
          </div>
        )
      ) : errors ? (
        <div>{errors}</div>
      ) : (
        <LogBlock
          {...{
            data: messages,
            follow,
            enableSearch,
            selectableLines,
            title,
            noScrollBar,
            solid,
            maxLines,
            fontSize,
            actionComponent: (
              <div className="flex gap-xl">
                <div
                  onClick={() => {
                    if (!fullScreen) {
                      setClassName('z-50');
                    } else {
                      removeClassName('z-50');
                    }
                    setFullScreen((s) => !s);
                  }}
                  className="flex items-center justify-center font-bold text-xl cursor-pointer select-none active:translate-y-[1px] transition-all"
                >
                  {fullScreen ? (
                    <ArrowsIn size={16} />
                  ) : (
                    <ArrowsOut size={16} />
                  )}
                </div>
                {actionComponent}
              </div>
            ),
            width: fullScreen ? '100vw' : width,
            height: fullScreen ? '100vh' : height,
            hideLines,
            language,
          }}
        />
      )}
    </div>
  );
};

export default HighlightJsLog;
