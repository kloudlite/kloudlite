import hljs from 'highlight.js';
import { ViewportList } from 'react-viewport-list';
import * as sock from 'websocket';
import { ReactNode, useCallback, useEffect, useRef, useState } from 'react';
import classNames from 'classnames';
import { v4 as uuid } from 'uuid';
import axios from 'axios';
import Anser from 'anser';
import { ArrowsIn, ArrowsOut, List } from '@jengaicons/react';
import { useLog } from '~/root/lib/client/hooks/use-log';

const padLeadingZeros = (num: number, size: number) => {
  let s = `${num}`;
  while (s.length < size) s = `0${s}`;
  return s;
};

const getIndicesOf = (sourceStr: string, searchStr: string) => {
  const maxMatch = 20;
  let totalMatched = 0;

  if (!searchStr) return [];
  const pat = new RegExp(searchStr, 'gi');
  let found = pat.exec(sourceStr);
  const res = [];
  while (found) {
    totalMatched += 1;
    res.push([found.index, pat.lastIndex]);
    if (pat.lastIndex === sourceStr.length) break;
    found = pat.exec(sourceStr);
    if (totalMatched >= maxMatch) {
      console.log(`more than ${maxMatch} found`);
      break;
    }
  }
  return res;
};

const useSearch = (
  {
    data,
    searchText,
  }: {
    data: string[];
    searchText: string;
  },
  dependency: any[] = []
) => {
  return useCallback(() => {
    if (!searchText)
      return data.map((item, index) => ({
        line: item,
        searchInf: {
          match: [],
          idx: index,
        },
      }));
    return data
      .map((item, index) => {
        let sResult: number[][] = [];
        try {
          sResult = getIndicesOf(item, searchText);
        } catch (err) {
          console.error(err);
        }
        return {
          line: item,
          searchInf: {
            match: sResult?.length ? sResult : [],
            idx: index,
          },
        };
      })
      .filter((a) => a.searchInf.match);
  }, dependency)();
};

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
  useLog(inlineData);
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
  searchInf: {
    idx: number;
  };
  fontSize: number;
  lines: number;
  dark: boolean;
}
const LineNumber = ({ searchInf, fontSize, lines, dark }: ILineNumber) => {
  const ref = useRef(null);
  const [data, setData] = useState(() =>
    padLeadingZeros(searchInf.idx + 1, `${lines}`.length)
  );

  useEffect(() => {
    setData(padLeadingZeros(searchInf.idx + 1, `${lines}`.length));
  }, [lines, searchInf]);
  return (
    <code
      key={`ind+${searchInf.idx}`}
      className="inline-flex gap-4 items-center whitespace-pre"
      ref={ref}
    >
      <span className="hljs flex sticky left-0" style={{ fontSize }}>
        <HighlightIt
          {...{
            enableHL: true,
            inlineData: data,
            language: 'accesslog',
            className: classNames('border-b px-2', {
              'bg-gray-800 border-gray-700 ': dark,
              'bg-gray-200 border-gray-300 ': !dark,
            }),
          }}
        />
        <div className="hljs" style={{ width: fontSize / 2 }} />
      </span>
    </code>
  );
};

interface IFilterdHighlightIt {
  searchInf?: {
    idx: number;
    match: number[][];
  };
  inlineData: string;
  className?: string;
  language: string;
  dark: boolean;
}

const FilterdHighlightIt = ({
  searchInf,
  inlineData = '',
  className = '',
  language,
  dark,
}: IFilterdHighlightIt) => {
  if (!inlineData) {
    // eslint-disable-next-line no-param-reassign
    inlineData = ' ';
  }
  const [res, setRes] = useState<JSX.Element[]>([]);

  useEffect(() => {
    // TODO: multi match
    (async () => {
      if (searchInf?.match.length) {
        setRes(
          searchInf.match.reduce(
            // @ts-ignore
            (acc, curr, index) => {
              return {
                cursor: curr[1],
                res: [
                  ...acc.res,
                  ...(inlineData.slice(acc.cursor, curr[0])
                    ? [
                        <HighlightIt
                          language={language}
                          key={
                            searchInf.idx +
                            inlineData.slice(acc.cursor, curr[0])
                          }
                          // key={uuid()}
                          inlineData={inlineData.slice(acc.cursor, curr[0])}
                          className={className}
                        />,
                      ]
                    : []),
                  <span
                    // key={searchInf.idx + inlineData.slice(curr[0], curr[1])}
                    key={uuid()}
                    className={classNames(className, ' rounded-sm', {
                      'bg-surface-warning-default text-text-warning': dark,
                      'bg-surface-critical-default text-text-critical': !dark,
                    })}
                  >
                    {inlineData.slice(curr[0], curr[1]) || ' '}
                  </span>,
                  ...[
                    index === searchInf.match.length - 1 &&
                      curr[1] !== index && (
                        <HighlightIt
                          key={searchInf.idx + inlineData.slice(curr[1])}
                          // key={uuid()}
                          inlineData={inlineData.slice(curr[1])}
                          className={className}
                          language={language}
                        />
                      ),
                  ],
                ],
              };
            },
            {
              cursor: 0,
              res: [],
            }
            // @ts-ignore
          ).res
        );
      } else {
        setRes([
          <HighlightIt
            key={inlineData}
            inlineData={inlineData}
            className={className}
            language={language}
          />,
        ]);
      }
    })();
  }, [searchInf]);

  return <div className="whitespace-pre">{res}</div>;
};

interface IWithSearchHighlightIt {
  inlineData: string;
  className?: string;
  searchText: string;
  language: string;
  dark: boolean;
}

const WithSearchHighlightIt = ({
  inlineData = '',
  className = '',
  searchText = '',
  language,
  dark,
}: IWithSearchHighlightIt) => {
  const x = useSearch(
    {
      data: [inlineData],
      searchText,
    },
    [inlineData, searchText]
  );

  return (
    <FilterdHighlightIt
      {...{
        inlineData,
        className,
        dark,
        language,
        ...(x.length ? { searchInf: x[0].searchInf } : {}),
      }}
    />
  );
};

interface ILogLine {
  searchInf: {
    idx: number;
    match: number[][];
  };
  line: string;
  fontSize: number;
  selectableLines: boolean;
  showAll: boolean;
  searchText: string;
  language: string;
  dark: boolean;
  lines: number;
  hideLines?: boolean;
}

const LogLine = ({
  searchInf,
  line,
  fontSize,
  selectableLines,
  showAll,
  searchText,
  language,
  dark,
  lines,
  hideLines,
}: ILogLine) => {
  return (
    <code
      className={classNames(
        'flex gap-4 items-center whitespace-pre border-b border-transparent',
        {
          'hover:bg-gray-800': selectableLines && dark,
          'hover:bg-gray-200': selectableLines && !dark,
        }
      )}
      style={{
        fontSize,
        paddingLeft: fontSize / 2,
        paddingRight: fontSize / 2,
      }}
    >
      {!hideLines && (
        <LineNumber
          searchInf={searchInf}
          lines={lines}
          fontSize={fontSize}
          dark={dark}
        />
      )}
      {showAll ? (
        <WithSearchHighlightIt
          {...{
            inlineData: line,
            searchText,
            language,
            dark,
          }}
        />
      ) : (
        <FilterdHighlightIt
          {...{
            inlineData: line,
            dark,
            searchInf,
            language,
          }}
        />
      )}
    </code>
  );
};

interface ILogBlock {
  data: string;
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
  dark: boolean;
  solid: boolean;
}

const LogBlock = ({
  data = '',
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
  dark,
  solid,
}: ILogBlock) => {
  const lines = data.split('\n');

  const [searchText, setSearchText] = useState('');

  const x = useSearch(
    {
      // eslint-disable-next-line no-nested-ternary
      data: maxLines
        ? lines.length >= maxLines
          ? lines.slice(lines.length - maxLines)
          : lines
        : lines,
      searchText,
    },
    [data, searchText, maxLines]
  );

  const [showAll, setShowAll] = useState(false);
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
      className={classNames('hljs p-2 flex flex-col gap-2 h-full', {
        border: !dark,
        'rounded-md': !solid,
      })}
    >
      <div className="flex justify-between px-2 items-center border-b border-gray-500 pb-3">
        <div className="">
          {data ? title : 'No logs generated in last 24 hours'}
        </div>

        <div className="flex items-center gap-3">
          {actionComponent}
          {enableSearch && (
            <form
              className="flex gap-3 items-center text-sm"
              onSubmit={(e) => {
                e.preventDefault();
                setShowAll((s) => !s);
              }}
            >
              <input
                className="bg-transparent border border-gray-400 rounded-md px-2 py-0.5 w-[10rem]"
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
                    'text-gray-200': !showAll,
                    'text-gray-600': showAll,
                  })}
                >
                  <List color="currentColor" />
                </span>
              </div>
              <code
                className={classNames('text-xs font-bold', {
                  ...(!dark
                    ? {
                        'text-gray-600': (searchText ? x.length : 0) !== 0,
                        'text-gray-200': (searchText ? x.length : 0) === 0,
                      }
                    : {
                        'text-gray-200': (searchText ? x.length : 0) !== 0,
                        'text-gray-600': (searchText ? x.length : 0) === 0,
                      }),
                })}
              >
                {x.reduce(
                  (acc, { searchInf }) => acc + (searchInf.match?.length || 0),
                  0
                )}{' '}
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
          'hljs-log-scrollbar-night': !noScrollBar && dark,
          'hljs-log-scrollbar-dary': !noScrollBar && !dark,
        })}
      >
        <div className="flex flex-1 h-full">
          <div
            className="flex-1 flex flex-col pb-8"
            style={{ lineHeight: `${fontSize * 1.5}px` }}
            ref={ref}
          >
            <ViewportList
              items={x.filter((i) => {
                if (showAll) return true;
                if (!searchText) return true;
                return i.searchInf.match.length;
              })}
            >
              {({ line, searchInf }) => {
                return (
                  <LogLine
                    key={searchInf.idx}
                    {...{
                      lines: x.length,
                      dark,
                      searchInf,
                      line,
                      fontSize,
                      selectableLines,
                      showAll,
                      searchText,
                      language,
                      hideLines,
                    }}
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
  websocketOptions?: {
    formatMessage: (message: string) => string;
  };
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
  dark?: boolean;
  language?: string;
  solid?: boolean;
  className?: string;
}

const HighlightJsLog = ({
  websocket = false,
  websocketOptions = {
    formatMessage: (_: string): string => '',
  },
  follow = true,
  url = '',
  text,
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
  dark = true,
  language = 'accesslog',
  solid = false,
  className = '',
}: IHighlightJsLog) => {
  const [data, setData] = useState(text || '');
  const { formatMessage } = websocketOptions;
  const [isLoading, setIsLoading] = useState(false);
  const [fullScreen, setFullScreen] = useState(false);

  useEffect(() => {
    setData(text || '');
  }, [text]);

  useEffect(() => {
    (async () => {
      if (!url || websocket) return;
      setIsLoading(true);
      try {
        const d = await axios({
          url,
          method: 'GET',
        });
        setData((d.data || '').trim());
      } catch (err) {
        setData(
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
    if (!url || !websocket) return;

    let wsclient;
    setIsLoading(true);
    try {
      // eslint-disable-next-line new-cap
      wsclient = new sock.w3cwebsocket(url, '', '', {});
    } catch (err) {
      setIsLoading(false);
      setData(
        `${(err as Error).message}
An error occurred attempting to load the provided log.
Please check the URL and ensure it is reachable.
${url}`
      );
      return;
    }
    // wsclient.onopen = logger.log;
    // wsclient.onclose = logger.log;
    // wsclient.onerror = logger.log;

    wsclient.onmessage = (msg) => {
      try {
        const m = formatMessage ? formatMessage(msg.data.toString()) : msg;
        setData((s) => `${s}${m ? `\n${m}` : ''}`);
        setIsLoading(false);
      } catch (err) {
        console.log(err);
        setData("'Something went wrong! Please try again.'");
      }
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
          <div className="hljs p-2 rounded-md flex flex-col gap-2 items-center justify-center h-full">
            <code className="">
              <HighlightIt language={language} inlineData="Loading..." />
            </code>
          </div>
        )
      ) : (
        <LogBlock
          {...{
            data,
            follow,
            enableSearch,
            selectableLines,
            title,
            noScrollBar,
            solid,
            maxLines,
            fontSize,
            actionComponent: (
              <div className="flex gap-4">
                <div
                  onClick={() => setFullScreen((s) => !s)}
                  className="flex items-center justify-center font-bold text-xl cursor-pointer select-none active:translate-y-[1px] transition-all"
                >
                  {fullScreen ? <ArrowsIn /> : <ArrowsOut />}
                </div>
                {actionComponent}
              </div>
            ),
            width: fullScreen ? '100vw' : width,
            height: fullScreen ? '100vh' : height,
            hideLines,
            language,
            dark,
          }}
        />
      )}
    </div>
  );
};

export default HighlightJsLog;
